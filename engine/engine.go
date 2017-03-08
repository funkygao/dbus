// Package provides a plugin based pipeline engine that decouples Input/Filter/Output plugins.
package engine

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

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

	// REST exporter
	httpListener net.Listener
	httpServer   *http.Server
	httpRouter   *mux.Router
	httpPaths    []string

	projects map[string]*Project // TODO

	InputRunners  map[string]InputRunner
	inputWrappers map[string]*pluginWrapper

	FilterRunners  map[string]FilterRunner
	filterWrappers map[string]*pluginWrapper

	OutputRunners  map[string]OutputRunner
	outputWrappers map[string]*pluginWrapper

	top    *topology
	router *messageRouter

	inputRecycleChan  chan *Packet
	filterRecycleChan chan *Packet

	hostname string
	pid      int
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

	return &Engine{
		InputRunners:   make(map[string]InputRunner),
		inputWrappers:  make(map[string]*pluginWrapper),
		FilterRunners:  make(map[string]FilterRunner),
		filterWrappers: make(map[string]*pluginWrapper),
		OutputRunners:  make(map[string]OutputRunner),
		outputWrappers: make(map[string]*pluginWrapper),

		inputRecycleChan:  make(chan *Packet, globals.RecyclePoolSize),
		filterRecycleChan: make(chan *Packet, globals.RecyclePoolSize),

		top:    newTopology(),
		router: newMessageRouter(),

		projects:  make(map[string]*Project),
		httpPaths: make([]string, 0, 6),

		pid:      os.Getpid(),
		hostname: hostname,
	}
}

func (e *Engine) stopInputRunner(name string) {
	e.Lock()
	e.InputRunners[name] = nil
	e.Unlock()
}

// Project returns a Project by name.
func (e *Engine) Project(name string) *Project {
	p, present := e.projects[name]
	if !present {
		return nil
	}

	return p
}

// NewPacket is used for plugin Filter to generate new Packet.
// The generated Packet will use dedicated filter recycle chan.
func (e *Engine) NewPacket() *Packet {
	pack := <-e.filterRecycleChan
	pack.Reset()
	return pack
}

func (e *Engine) LoadConfigFile(fn string) *Engine {
	if fn == "" {
		panic("config file is required")
	}
	if _, err := os.Stat(fn); err != nil {
		panic(err)
	}
	cf, err := conf.Load(fn)
	if err != nil {
		panic(err)
	}

	e.Conf = cf
	Globals().Conf = cf

	// 'projects' section
	for i := 0; i < len(e.List("projects", nil)); i++ {
		section, err := e.Section(fmt.Sprintf("projects[%d]", i))
		if err != nil {
			panic(err)
		}

		project := &Project{}
		project.fromConfig(section)
		if _, present := e.projects[project.Name]; present {
			panic("dup project: " + project.Name)
		}
		e.projects[project.Name] = project
	}

	// 'plugins' section
	for i := 0; i < len(e.List("plugins", nil)); i++ {
		section, err := e.Section(fmt.Sprintf("plugins[%d]", i))
		if err != nil {
			panic(err)
		}

		e.loadPluginSection(section)
	}

	// 'topology' section
	e.top.load(e.Conf)

	if c, err := influxdb.NewConfig(cf.String("influx_addr", ""),
		cf.String("influx_db", "dbus"), "", "",
		cf.Duration("influx_tick", time.Minute)); err == nil {
		telemetry.Default = influxdb.New(metrics.DefaultRegistry, c)
	} else {
		log.Warn("telemetry disabled for: %s", err)
	}

	return e
}

func (e *Engine) loadPluginSection(section *conf.Conf) {
	pluginCommons := new(pluginCommons)
	pluginCommons.loadConfig(section)
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
		e.InputRunners[wrapper.name] = newInputRunner(wrapper.name, plugin.(Input), pluginCommons)
		e.inputWrappers[wrapper.name] = wrapper

		return
	}

	foRunner := newFORunner(wrapper.name, plugin, pluginCommons)
	matcher := newMatcher(section.StringList("match", nil), foRunner)
	foRunner.matcher = matcher

	switch pluginCategory {
	case "Filter":
		e.router.addFilterMatcher(matcher)
		e.FilterRunners[foRunner.name] = foRunner
		e.filterWrappers[foRunner.name] = wrapper

	case "Output":
		e.router.addOutputMatcher(matcher)
		e.OutputRunners[foRunner.name] = foRunner
		e.outputWrappers[foRunner.name] = wrapper
	}
}

func (e *Engine) ServeForever() {
	var (
		outputsWg = new(sync.WaitGroup)
		filtersWg = new(sync.WaitGroup)
		inputsWg  = new(sync.WaitGroup)
		routerWg  = new(sync.WaitGroup)

		globals = Globals()
		err     error
	)

	// setup signal handler first to avoid race condition
	// if Input terminates very soon, global.Shutdown will
	// not be able to trap it
	globals.sigChan = make(chan os.Signal)
	signal.Notify(globals.sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGUSR1, syscall.SIGUSR2)

	e.launchHttpServ()

	if telemetry.Default != nil {
		log.Trace("launching telemetry dumper...")

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

	log.Trace("initializing Packet pool with size=%d", globals.RecyclePoolSize)
	for i := 0; i < globals.RecyclePoolSize; i++ {
		inputPack := NewPacket(e.inputRecycleChan)
		e.inputRecycleChan <- inputPack

		filterPack := NewPacket(e.filterRecycleChan)
		e.filterRecycleChan <- filterPack
	}

	log.Trace("launching Watchdog with ticker=%s", globals.WatchdogTick)
	go e.runWatchdog(globals.WatchdogTick)

	log.Trace("launching Router...")
	routerWg.Add(1)
	go e.router.Start(routerWg)

	for _, project := range e.projects {
		log.Trace("launching Project %s...", project.Name)

		project.Start()
	}

	for _, inputRunner := range e.InputRunners {
		log.Trace("launching Input[%s]...", inputRunner.Name())

		inputsWg.Add(1)
		if err = inputRunner.start(e, inputsWg); err != nil {
			inputsWg.Done()
			panic(err)
		}
	}

	for !globals.Stopping {
		select {
		case sig := <-globals.sigChan:
			log.Info("Got signal %s", strings.ToUpper(sig.String()))

			switch sig {
			case syscall.SIGHUP:
				log.Info("Reloading...")
				observer.Publish(RELOAD, nil)

			case syscall.SIGINT, syscall.SIGTERM:
				log.Info("Engine shutdown...")
				globals.Stopping = true
				if telemetry.Default != nil {
					telemetry.Default.Stop()
				}

			case syscall.SIGUSR1:
				observer.Publish(SIGUSR1, nil)

			case syscall.SIGUSR2:
				observer.Publish(SIGUSR2, nil)
			}
		}
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
	log.Trace("all Inputs stopped")

	// ok, now we are sure no more inputs, but in route.inChan there
	// still may be filter injected packs and output not consumed packs
	// we must wait for all the packs to be consumed before shutdown

	for _, filterRunner := range e.FilterRunners {
		log.Trace("Stop message sent to %s", filterRunner.Name())
		e.router.removeFilterMatcher <- filterRunner.getMatcher()
	}
	filtersWg.Wait()
	if len(e.FilterRunners) > 0 {
		log.Trace("all Filters stopped")
	}

	for _, outputRunner := range e.OutputRunners {
		log.Trace("Stop message sent to %s", outputRunner.Name())
		e.router.removeOutputMatcher <- outputRunner.getMatcher()
	}
	outputsWg.Wait()
	log.Trace("all Outputs stopped")

	// stop router
	close(e.router.hub)
	routerWg.Wait()
	log.Trace("Router stopped")

	e.stopHttpServ()

	for _, project := range e.projects {
		project.Stop()
	}

	log.Info("shutdown complete")
}
