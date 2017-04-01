// +build !v2

package engine

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/funkygao/dbus/pkg/cluster"
	czk "github.com/funkygao/dbus/pkg/cluster/zk"
	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gafka/telemetry"
	"github.com/funkygao/gafka/telemetry/influxdb"
	"github.com/funkygao/go-metrics"
	"github.com/funkygao/golib/observer"
	conf "github.com/funkygao/jsconf"
	log "github.com/funkygao/log4go"
	"github.com/gorilla/mux"
)

var (
	_ PluginHelper = &Engine{}
)

// Engine is the pipeline engine of the data bus system which manages the core loop.
type Engine struct {
	sync.RWMutex

	// Engine will load json config file
	*conf.Conf

	participant cluster.Participant
	controller  cluster.Controller

	// API Server
	apiListener net.Listener
	apiServer   *http.Server
	apiRouter   *mux.Router
	httpPaths   []string

	// RPC Server
	rpcListener net.Listener
	rpcServer   *http.Server
	rpcRouter   *mux.Router

	// input plugin resources map
	irm   map[string][]cluster.Resource
	irmMu sync.Mutex

	InputRunners  map[string]*iRunner
	inputWrappers map[string]*pluginWrapper

	FilterRunners  map[string]FilterRunner
	filterWrappers map[string]*pluginWrapper

	OutputRunners  map[string]OutputRunner
	outputWrappers map[string]*pluginWrapper

	router *Router

	inputRecycleChans map[string]chan *Packet
	filterRecycleChan chan *Packet

	hostname string
	pid      int
	stopper  chan struct{}
	epoch    int
}

func New(globals *GlobalConfig) *Engine {
	if globals == nil {
		globals = DefaultGlobals()
	}
	Globals = func() *GlobalConfig {
		return globals
	}

	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	// make the participant
	localIP, err := ctx.LocalIP()
	if err != nil {
		panic(err)
	}

	return &Engine{
		InputRunners:   make(map[string]*iRunner),
		inputWrappers:  make(map[string]*pluginWrapper),
		FilterRunners:  make(map[string]FilterRunner),
		filterWrappers: make(map[string]*pluginWrapper),
		OutputRunners:  make(map[string]OutputRunner),
		outputWrappers: make(map[string]*pluginWrapper),

		inputRecycleChans: make(map[string]chan *Packet),
		filterRecycleChan: make(chan *Packet, globals.FilterRecyclePoolSize),

		router: newRouter(),

		httpPaths: make([]string, 0, 6),

		pid:      os.Getpid(),
		hostname: hostname,
		stopper:  make(chan struct{}),
		participant: cluster.Participant{
			Endpoint: fmt.Sprintf("%s:%d", localIP.String(), globals.RPCPort),
			Weight:   runtime.NumCPU() * 100,
		},
	}
}

func (e *Engine) stopInputRunner(name string) {
	e.Lock()
	e.InputRunners[name] = nil
	e.Unlock()
}

// Leader returns the cluster leader RPC address.
func (e *Engine) Leader() string {
	return ""
}

// ClonePacket is used for plugin Filter to generate new Packet.
// The generated Packet will use dedicated filter recycle chan.
func (e *Engine) ClonePacket(p *Packet) *Packet {
	pack := <-e.filterRecycleChan
	pack.Reset()
	p.copyTo(pack)
	return pack
}

func (e *Engine) LoadConfig(path string) *Engine {
	if len(path) == 0 {
		// if no path provided, use the default zk
		path = fmt.Sprintf("%s%s", ctx.ZoneZkAddrs(ctx.DefaultZone()), DbusConfZnode)
	}

	zkSvr, realPath := parseConfigPath(path)
	var (
		cf  *conf.Conf
		err error
	)
	if len(zkSvr) == 0 {
		// from file system
		cf, err = conf.Load(realPath)
	} else {
		// from zookeeper
		cf, err = conf.Load(realPath, conf.WithZkSvr(zkSvr))
		if err != nil {
			err = fmt.Errorf("%s %v", path, err)
		}
	}
	if err != nil {
		panic(err)
	}

	e.Conf = cf
	Globals().Conf = cf

	if Globals().ClusterEnabled {
		e.controller = czk.NewController(zkSvr, e.participant, cluster.StrategyRoundRobin, e.onControllerRebalance)
	}

	// 'plugins' section
	var names = make(map[string]struct{})
	for i := 0; i < len(e.List("plugins", nil)); i++ {
		section, err := e.Section(fmt.Sprintf("plugins[%d]", i))
		if err != nil {
			panic(err)
		}

		name := e.loadPluginSection(section)
		if _, duplicated := names[name]; duplicated {
			// router.metrics will be bad with dup name
			panic("duplicated plugin name: " + name)
		}
		names[name] = struct{}{}
	}

	if c, err := influxdb.NewConfig(cf.String("influx_addr", ""),
		cf.String("influx_db", "dbus"), "", "",
		cf.Duration("influx_tick", time.Minute)); err == nil {
		telemetry.Default = influxdb.New(metrics.DefaultRegistry, c)
	}

	return e
}

func (e *Engine) loadPluginSection(section *conf.Conf) (name string) {
	pluginCommons := new(pluginCommons)
	pluginCommons.loadConfig(section)
	name = pluginCommons.name
	if pluginCommons.disabled {
		log.Warn("%s disabled", pluginCommons.name)

		return
	}

	wrapper := &pluginWrapper{
		name:          pluginCommons.name,
		configCreator: func() *conf.Conf { return section },
	}
	var ok bool
	if wrapper.pluginCreator, ok = availablePlugins[pluginCommons.class]; !ok {
		panic("unknown plugin type: " + pluginCommons.class)
	}

	pluginType := pluginTypeRegex.FindStringSubmatch(pluginCommons.class)
	if len(pluginType) < 2 {
		panic("invalid plugin type: " + pluginCommons.class)
	}

	plugin := wrapper.Create()
	pluginCategory := pluginType[1]
	if pluginCategory == "Input" {
		e.inputRecycleChans[wrapper.name] = make(chan *Packet, Globals().InputRecyclePoolSize)
		e.InputRunners[wrapper.name] = newInputRunner(plugin.(Input), pluginCommons)
		e.inputWrappers[wrapper.name] = wrapper
		e.router.metrics.m[wrapper.name] = metrics.NewRegisteredMeter(wrapper.name, metrics.DefaultRegistry)
		return
	}

	foRunner := newFORunner(plugin, pluginCommons)
	matcher := newMatcher(section.StringList("match", nil), foRunner)
	foRunner.matcher = matcher

	switch pluginCategory {
	case "Filter":
		e.router.addFilterMatcher(matcher)
		e.FilterRunners[foRunner.Name()] = foRunner
		e.filterWrappers[foRunner.Name()] = wrapper
		e.router.metrics.m[wrapper.name] = metrics.NewRegisteredMeter(wrapper.name, metrics.DefaultRegistry)

	case "Output":
		e.router.addOutputMatcher(matcher)
		e.OutputRunners[foRunner.Name()] = foRunner
		e.outputWrappers[foRunner.Name()] = wrapper
		e.router.metrics.m[wrapper.name] = metrics.NewRegisteredMeter(wrapper.name, metrics.DefaultRegistry)

	default:
		panic("unknown plugin: " + pluginCategory)
	}

	return
}

func (e *Engine) ServeForever() (ret error) {
	var (
		outputsWg = new(sync.WaitGroup)
		filtersWg = new(sync.WaitGroup)
		inputsWg  = new(sync.WaitGroup)
		routerWg  = new(sync.WaitGroup)

		globals = Globals()
		err     error
	)

	if telemetry.Default == nil {
		log.Info("engine starting, with telemetry disabled...")
	} else {
		log.Info("engine starting...")
	}

	// setup signal handler first to avoid race condition
	// if Input terminates very soon, global.Shutdown will
	// not be able to trap it
	globals.sigChan = make(chan os.Signal)
	signal.Notify(globals.sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)

	e.launchAPIServer()
	if globals.ClusterEnabled {
		e.launchRPCServer()

		if err = e.controller.Start(); err != nil {
			panic(err)
		}
		log.Info("[%s] controller started", e.participant)
	}

	if telemetry.Default != nil {
		log.Info("launching telemetry dumper...")

		go func() {
			if err := telemetry.Default.Start(); err != nil {
				log.Error("telemetry[%s]: %s", telemetry.Default.Name(), err)
			}
		}()
	}

	for _, outputRunner := range e.OutputRunners {
		log.Trace("launching Output[%s]...", outputRunner.Name())

		outputsWg.Add(1)
		if err = outputRunner.start(e, outputsWg); err != nil {
			panic(err)
		}
	}

	for _, filterRunner := range e.FilterRunners {
		log.Trace("launching Filter[%s]...", filterRunner.Name())

		filtersWg.Add(1)
		if err = filterRunner.start(e, filtersWg); err != nil {
			panic(err)
		}
	}

	for inputName := range e.inputRecycleChans {
		log.Info("building Input[%s] Packet pool with size=%d", inputName, globals.InputRecyclePoolSize)

		for i := 0; i < globals.InputRecyclePoolSize; i++ {
			inputPack := newPacket(e.inputRecycleChans[inputName])
			e.inputRecycleChans[inputName] <- inputPack
		}
	}

	log.Info("building Filter Packet pool with size=%d", globals.FilterRecyclePoolSize)
	for i := 0; i < globals.FilterRecyclePoolSize; i++ {
		filterPack := newPacket(e.filterRecycleChan)
		e.filterRecycleChan <- filterPack
	}

	log.Info("launching Watchdog with ticker=%s", globals.WatchdogTick)
	go e.runWatchdog(globals.WatchdogTick)

	routerWg.Add(1)
	go e.router.Start(routerWg)

	for _, inputRunner := range e.InputRunners {
		log.Trace("launching Input[%s]...", inputRunner.Name())

		inputsWg.Add(1)
		if err = inputRunner.start(e, inputsWg); err != nil {
			inputsWg.Done()
			panic(err)
		}
	}

	cfChanged := make(chan *conf.Conf)
	go e.Conf.Watch(time.Second*10, e.stopper, cfChanged)

	for !globals.Stopping {
		select {
		case <-cfChanged:
			log.Info("%s updated, closing...", e.Conf.ConfPath())
			globals.Stopping = true

		case sig := <-globals.sigChan:
			log.Info("Got signal %s", strings.ToUpper(sig.String()))

			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				log.Info("shutdown...")
				globals.Stopping = true
				ret = ErrQuitingSigal

			case syscall.SIGUSR1:
				observer.Publish(SIGUSR1, nil)

			case syscall.SIGUSR2:
				observer.Publish(SIGUSR2, nil)
			}
		}
	}

	close(e.stopper)

	if telemetry.Default != nil {
		telemetry.Default.Stop()
	}

	e.Lock()
	for _, inputRunner := range e.InputRunners {
		if inputRunner == nil {
			// the Input plugin already exit
			continue
		}

		log.Trace("Stop message sent to %s", inputRunner.Name())
		inputRunner.Input().Stop(inputRunner)
	}
	e.Unlock()
	inputsWg.Wait() // wait for all inputs done
	log.Info("all Inputs stopped")

	// ok, now we are sure no more inputs, but in route.inChan there
	// still may be filter injected packs and output not consumed packs
	// we must wait for all the packs to be consumed before shutdown

	for _, filterRunner := range e.FilterRunners {
		log.Trace("Stop message sent to %s", filterRunner.Name())
		e.router.removeFilterMatcher <- filterRunner.getMatcher()
	}
	filtersWg.Wait()
	if len(e.FilterRunners) > 0 {
		log.Info("all Filters stopped")
	}

	for _, outputRunner := range e.OutputRunners {
		log.Trace("Stop message sent to %s", outputRunner.Name())
		e.router.removeOutputMatcher <- outputRunner.getMatcher()
	}
	outputsWg.Wait()
	log.Info("all Outputs stopped")

	e.router.Stop()
	routerWg.Wait()
	log.Info("Router stopped")

	e.stopAPIServer()

	if globals.ClusterEnabled {
		e.stopRPCServer()

		if err = e.controller.Stop(); err != nil {
			log.Error("%v", err)
		}
	}

	if ret != nil {
		log.Info("shutdown complete: %s!", ret)
	} else {
		log.Info("shutdown complete!")
	}

	return
}
