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
	"sync/atomic"
	"syscall"
	"time"

	"github.com/funkygao/gafka/telemetry"
	"github.com/funkygao/gafka/telemetry/influxdb"
	"github.com/funkygao/go-metrics"
	"github.com/funkygao/golib/gofmt"
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

	// PipelinePack supply for Input plugins.
	inputRecycleChan chan *PipelinePack

	// PipelinePack supply for Filter plugins.
	filterRecycleChan chan *PipelinePack

	hostname string
	pid      int
}

func New(globals *GlobalConfig) (this *Engine) {
	this = new(Engine)

	if globals == nil {
		globals = DefaultGlobals()
	}
	Globals = func() *GlobalConfig {
		return globals
	}

	this.InputRunners = make(map[string]InputRunner)
	this.inputWrappers = make(map[string]*pluginWrapper)
	this.FilterRunners = make(map[string]FilterRunner)
	this.filterWrappers = make(map[string]*pluginWrapper)
	this.OutputRunners = make(map[string]OutputRunner)
	this.outputWrappers = make(map[string]*pluginWrapper)

	this.inputRecycleChan = make(chan *PipelinePack, globals.RecyclePoolSize)
	this.filterRecycleChan = make(chan *PipelinePack, globals.RecyclePoolSize)

	this.projects = make(map[string]*Project)
	this.httpPaths = make([]string, 0, 6)

	this.top = newTopology()
	this.router = newMessageRouter()

	this.hostname, _ = os.Hostname()
	this.pid = os.Getpid()

	return this
}

func (this *Engine) stopInputRunner(name string) {
	this.Lock()
	this.InputRunners[name] = nil
	this.Unlock()
}

func (this *Engine) Engine() *Engine {
	return this
}

func (this *Engine) Project(name string) *Project {
	p, present := this.projects[name]
	if !present {
		return nil
	}

	return p
}

func (this *Engine) LoadConfigFile(fn string) *Engine {
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

	this.Conf = cf
	Globals().Conf = cf

	// 'projects' section
	for i := 0; i < len(this.List("projects", nil)); i++ {
		section, err := this.Section(fmt.Sprintf("projects[%d]", i))
		if err != nil {
			panic(err)
		}

		project := &Project{}
		project.fromConfig(section)
		if _, present := this.projects[project.Name]; present {
			panic("dup project: " + project.Name)
		}
		this.projects[project.Name] = project
	}

	// 'plugins' section
	for i := 0; i < len(this.List("plugins", nil)); i++ {
		section, err := this.Section(fmt.Sprintf("plugins[%d]", i))
		if err != nil {
			panic(err)
		}

		this.loadPluginSection(section)
	}

	// 'topology' section
	this.top.load(this.Conf)

	if c, err := influxdb.NewConfig(cf.String("influx_addr", ""),
		cf.String("influx_db", "dbus"), "", "",
		cf.Duration("influx_tick", time.Minute)); err == nil {
		telemetry.Default = influxdb.New(metrics.DefaultRegistry, c)
	} else {
		log.Warn("telemetry disabled for: %s", err)
	}

	return this
}

func (this *Engine) loadPluginSection(section *conf.Conf) {
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
		this.InputRunners[wrapper.name] = newInputRunner(wrapper.name, plugin.(Input), pluginCommons)
		this.inputWrappers[wrapper.name] = wrapper

		return
	}

	foRunner := newFORunner(wrapper.name, plugin, pluginCommons)
	matcher := newMatcher(section.StringList("match", nil), foRunner)
	foRunner.matcher = matcher

	switch pluginCategory {
	case "Filter":
		this.router.addFilterMatcher(matcher)
		this.FilterRunners[foRunner.name] = foRunner
		this.filterWrappers[foRunner.name] = wrapper

	case "Output":
		this.router.addOutputMatcher(matcher)
		this.OutputRunners[foRunner.name] = foRunner
		this.outputWrappers[foRunner.name] = wrapper
	}
}

func (this *Engine) ServeForever() {
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

	this.launchHttpServ()

	if telemetry.Default != nil {
		log.Trace("launching telemetry dumper...")

		go func() {
			if err := telemetry.Default.Start(); err != nil {
				log.Error("telemetry[%s]: %s", telemetry.Default.Name(), err)
			}
		}()
	}

	for _, runner := range this.OutputRunners {
		log.Trace("launching Output[%s]...", runner.Name())

		outputsWg.Add(1)
		if err = runner.start(this, outputsWg); err != nil {
			panic(err)
		}
	}

	for _, runner := range this.FilterRunners {
		log.Trace("launching Filter[%s]...", runner.Name())

		filtersWg.Add(1)
		if err = runner.start(this, filtersWg); err != nil {
			panic(err)
		}
	}

	log.Trace("initializing PipelinePack pools with size=%d", globals.RecyclePoolSize)
	for i := 0; i < globals.RecyclePoolSize; i++ {
		inputPack := NewPipelinePack(this.inputRecycleChan)
		this.inputRecycleChan <- inputPack

		filterPack := NewPipelinePack(this.filterRecycleChan)
		this.filterRecycleChan <- filterPack
	}

	// check if we have enough recycle pool reservation
	go func() {
		t := time.NewTicker(globals.WatchdogTick)
		defer t.Stop()

		var inputPoolSize, filterPoolSize int

		for range t.C {
			inputPoolSize = len(this.inputRecycleChan)
			filterPoolSize = len(this.filterRecycleChan)
			if inputPoolSize == 0 || filterPoolSize == 0 {
				log.Trace("Recycle pool reservation: [input]%d [filter]%d", inputPoolSize, filterPoolSize)
			}
		}
	}()

	log.Trace("launching Router...")
	routerWg.Add(1)
	go this.router.Start(routerWg)

	for _, project := range this.projects {
		log.Trace("launching Project %s...", project.Name)

		project.Start()
	}

	for _, runner := range this.InputRunners {
		log.Trace("launching Input[%s]...", runner.Name())

		inputsWg.Add(1)
		if err = runner.start(this, inputsWg); err != nil {
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

	this.Lock()
	for _, runner := range this.InputRunners {
		if runner == nil {
			// this Input plugin already exit
			continue
		}

		log.Trace("Stop message sent to %s", runner.Name())
		runner.Input().Stop(runner)
	}
	this.Unlock()
	inputsWg.Wait() // wait for all inputs done
	log.Trace("all Inputs stopped")

	// ok, now we are sure no more inputs, but in route.inChan there
	// still may be filter injected packs and output not consumed packs
	// we must wait for all the packs to be consumed before shutdown

	for _, runner := range this.FilterRunners {
		log.Trace("Stop message sent to %s", runner.Name())
		this.router.removeFilterMatcher <- runner.getMatcher()
	}
	filtersWg.Wait()
	if len(this.FilterRunners) > 0 {
		log.Trace("all Filters stopped")
	}

	for _, runner := range this.OutputRunners {
		log.Trace("Stop message sent to %s", runner.Name())
		this.router.removeOutputMatcher <- runner.getMatcher()
	}
	outputsWg.Wait()
	log.Trace("all Outputs stopped")

	// stop router
	close(this.router.hub)
	routerWg.Wait()
	log.Trace("Router stopped")

	this.stopHttpServ()

	for _, project := range this.projects {
		project.Stop()
	}

	log.Info("shutdown with input:%s, dispatched:%s",
		gofmt.Comma(atomic.LoadInt64(&this.router.stats.TotalInputMsgN)),
		gofmt.Comma(atomic.LoadInt64(&this.router.stats.TotalProcessedMsgN)))
}
