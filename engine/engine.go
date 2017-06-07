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

	"github.com/funkygao/dbus"
	"github.com/funkygao/dbus/pkg/cluster"
	czk "github.com/funkygao/dbus/pkg/cluster/zk"
	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gafka/telemetry"
	"github.com/funkygao/gafka/telemetry/influxdb"
	"github.com/funkygao/gafka/zk"
	"github.com/funkygao/go-metrics"
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
	*conf.Conf

	zkSvr       string
	participant cluster.Participant
	controller  cluster.Controller
	epoch       int // cache of latest cluster leader epoch

	// API Server
	apiListener net.Listener
	apiServer   *http.Server
	apiRouter   *mux.Router

	// RPC Server
	rpcListener net.Listener
	rpcServer   *http.Server
	rpcRouter   *mux.Router

	// input plugin resources map
	irm   map[string][]cluster.Resource
	irmMu sync.Mutex

	// dataflow router
	router *Router

	InputRunners  map[string]*iRunner
	inputWrappers map[string]*pluginWrapper

	FilterRunners  map[string]FilterRunner
	filterWrappers map[string]*pluginWrapper

	OutputRunners  map[string]OutputRunner
	outputWrappers map[string]*pluginWrapper

	inputRecycleChans map[string]chan *Packet
	filterRecycleChan chan *Packet

	hostname      string
	pid           int
	stopper       chan struct{}
	shutdown      chan struct{}
	pluginPanicCh chan error
}

// New creates an engine.
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
		pid:           os.Getpid(),
		hostname:      hostname,
		stopper:       make(chan struct{}),
		shutdown:      make(chan struct{}),
		pluginPanicCh: make(chan error),

		router: newRouter(),

		InputRunners:   make(map[string]*iRunner),
		inputWrappers:  make(map[string]*pluginWrapper),
		FilterRunners:  make(map[string]FilterRunner),
		filterWrappers: make(map[string]*pluginWrapper),
		OutputRunners:  make(map[string]OutputRunner),
		outputWrappers: make(map[string]*pluginWrapper),

		inputRecycleChans: make(map[string]chan *Packet),
		filterRecycleChan: make(chan *Packet, globals.FilterRecyclePoolSize),

		participant: cluster.Participant{
			Endpoint: fmt.Sprintf("%s:%d", localIP.String(), globals.RPCPort),
			Weight:   runtime.NumCPU() * 100,
			State:    cluster.StateOnline,
			Revision: dbus.Revision,
			APIPort:  globals.APIPort,
		},
	}
}

// ClonePacket is used for plugin Filter to generate new Packet: copy on write.
// The generated Packet will use dedicated filter recycle chan.
func (e *Engine) ClonePacket(p *Packet) *Packet {
	pack := <-e.filterRecycleChan
	pack.Reset()
	p.copyTo(pack)
	return pack
}

// ClusterManager returns the cluster manager.
// If cluster is disabled, returns nil.
func (e *Engine) ClusterManager() cluster.Manager {
	if Globals().ClusterEnabled {
		return e.controller.(cluster.Manager)
	}

	return nil
}

// SubmitDAG submits a DAG configuration to the engine.
func (e *Engine) SubmitDAG(cf *conf.Conf) *Engine {
	return e.loadConfig(cf)
}

func (e *Engine) loadConfig(cf *conf.Conf) *Engine {
	e.Conf = cf
	Globals().Conf = cf

	// 'plugins' section
	var pluginNames = make(map[string]struct{})
	for i := 0; i < len(e.List("plugins", nil)); i++ {
		section, err := e.Section(fmt.Sprintf("plugins[%d]", i))
		if err != nil {
			panic(err)
		}

		name := e.loadPluginSection(section)
		if _, duplicated := pluginNames[name]; duplicated {
			panic("duplicated plugin name: " + name)
		}
		pluginNames[name] = struct{}{}
	}

	// influxdb related section
	if c, err := influxdb.NewConfig(cf.String("influx_addr", ""),
		cf.String("influx_db", "dbus"), "", "",
		cf.Duration("influx_tick", time.Minute)); err == nil {
		telemetry.Default = influxdb.New(metrics.DefaultRegistry, c)
	}

	return e
}

// LoadFrom load the configuration by location.
// The location can be empty: use default zk zone /dbus/conf.
// If config is stored on file, the loc arg is file path.
// If config is stored on zookeeper, the loc arg is like localhost:2181/foo/bar.
func (e *Engine) LoadFrom(loc string) *Engine {
	if len(loc) == 0 {
		// if no location provided, use the default zk
		loc = fmt.Sprintf("%s%s", ctx.ZoneZkAddrs(Globals().Zone), zk.DbusConfig(Globals().Cluster))
	}

	zkSvr, realPath := parseConfigPath(loc)
	var (
		cf  *conf.Conf
		err error
	)
	if len(zkSvr) == 0 {
		// from file system
		cf, err = conf.Load(realPath)
	} else {
		// from zookeeper
		e.zkSvr = zkSvr
		cf, err = conf.Load(realPath, conf.WithZkSvr(zkSvr))
		if err != nil {
			err = fmt.Errorf("%s %v", loc, err)
		}
	}
	if err != nil {
		panic(err)
	}

	return e.loadConfig(cf)
}

func (e *Engine) loadPluginSection(section *conf.Conf) string {
	pluginCommons := new(pluginCommons)
	pluginCommons.loadConfig(section)

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
		e.InputRunners[wrapper.name] = newInputRunner(plugin.(Input), pluginCommons, e.pluginPanicCh)
		e.inputWrappers[wrapper.name] = wrapper
		e.router.metrics.m[wrapper.name] = metrics.NewRegisteredMeter(wrapper.name, metrics.DefaultRegistry)
		return pluginCommons.name
	}

	foRunner := newFORunner(plugin, pluginCommons, e.pluginPanicCh)
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

	return pluginCommons.name
}

func (e *Engine) Shutdown() {
	close(e.shutdown)
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

	log.Trace("engine starting...")

	if globals.ClusterEnabled {
		e.controller = czk.NewController(e.zkSvr, globals.Cluster, e.participant, cluster.StrategyRoundRobin, e.leaderRebalance)
	}

	// setup signal handler first to avoid race condition
	// if Input terminates very fast, global.Shutdown will not be able to trap it
	globals.sigChan = make(chan os.Signal)
	signal.Notify(globals.sigChan, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)

	e.launchAPIServer()

	if telemetry.Default != nil {
		go func() {
			if err := telemetry.Default.Start(); err != nil {
				log.Error("telemetry[%s]: %s", telemetry.Default.Name(), err)
			}
		}()
	}

	for _, outputRunner := range e.OutputRunners {
		log.Debug("launching Output[%s]...", outputRunner.Name())

		outputsWg.Add(1)
		outputRunner.forkAndRun(e, outputsWg)
	}

	for _, filterRunner := range e.FilterRunners {
		log.Debug("launching Filter[%s]...", filterRunner.Name())

		filtersWg.Add(1)
		filterRunner.forkAndRun(e, filtersWg)
	}

	for inputName := range e.inputRecycleChans {
		for i := 0; i < globals.InputRecyclePoolSize; i++ {
			inputPack := newPacket(e.inputRecycleChans[inputName])
			e.inputRecycleChans[inputName] <- inputPack
		}
	}

	for i := 0; i < globals.FilterRecyclePoolSize; i++ {
		filterPack := newPacket(e.filterRecycleChan)
		e.filterRecycleChan <- filterPack
	}

	go e.runWatchdog(globals.WatchdogTick)

	routerWg.Add(1)
	go e.router.Start(routerWg)

	for _, inputRunner := range e.InputRunners {
		log.Debug("launching Input[%s]...", inputRunner.Name())

		inputsWg.Add(1)
		inputRunner.forkAndRun(e, inputsWg)
	}

	if globals.ClusterEnabled {
		e.launchRPCServer()

		log.Trace("[%s] participant starting...", e.participant)
		if err = e.controller.Start(); err != nil {
			panic(err)
		}
		go e.watchUpgrade(e.ClusterManager().Upgrade())

		log.Info("[%s] participant started", e.participant)
	} else {
		log.Info("cluster disabled")
	}

	configChanged := make(chan *conf.Conf)
	go e.Conf.Watch(time.Second*10, e.stopper, configChanged)

	log.Info("engine started")
	globals.stopping = false
	for !globals.stopping {
		select {
		case <-configChanged:
			log.Info("%s changed, shutdown...", e.Conf.ConfPath())
			globals.stopping = true

		case <-e.shutdown:
			log.Info("shutdown...")
			globals.stopping = true

		case sig := <-globals.sigChan:
			log.Info("Got signal %s", strings.ToUpper(sig.String()))

			switch sig {
			case syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP:
				log.Info("shutdown...")
				globals.stopping = true
				ret = ErrQuitingSigal
			}

		case ret = <-e.pluginPanicCh:
			log.Info("plugin panic, stopping...")
			globals.stopping = true
		}
	}

	close(e.stopper)

	inputsWg.Wait()

	e.router.Stop()
	routerWg.Wait()
	log.Info("Router stopped")

	filtersWg.Wait()
	outputsWg.Wait()

	for _, inputRunner := range e.InputRunners {
		inputRunner.Input().End(inputRunner)
		log.Trace("[%s] ended", inputRunner.Name())
	}

	log.Info("all %d plugins fully stopped", len(e.InputRunners)+len(e.FilterRunners)+len(e.OutputRunners))

	// now more packet flow, safe to close
	close(e.filterRecycleChan)
	for _, c := range e.inputRecycleChans {
		close(c)
	}

	// needn't close registered zkzones, they will not raise leakage

	if telemetry.Default != nil {
		telemetry.Default.Stop()
	}

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
