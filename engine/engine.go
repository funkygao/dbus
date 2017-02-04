package engine

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/funkygao/golib/gofmt"
	"github.com/funkygao/golib/observer"
	conf "github.com/funkygao/jsconf"
	"github.com/funkygao/pretty"
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

	projects map[string]*ConfProject // TODO

	InputRunners  map[string]InputRunner
	inputWrappers map[string]*pluginWrapper

	FilterRunners  map[string]FilterRunner
	filterWrappers map[string]*pluginWrapper

	OutputRunners  map[string]OutputRunner
	outputWrappers map[string]*pluginWrapper

	diagnosticTrackers map[string]*diagnosticTracker

	router *messageRouter
	stats  *engineStats

	// PipelinePack supply for Input plugins.
	inputRecycleChan chan *PipelinePack

	// PipelinePack supply for Filter plugins
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

	this.diagnosticTrackers = make(map[string]*diagnosticTracker)
	this.projects = make(map[string]*ConfProject)
	this.httpPaths = make([]string, 0, 6)

	this.router = newMessageRouter()
	this.stats = newEngineStats(this)

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

func (this *Engine) Project(name string) *ConfProject {
	p, present := this.projects[name]
	if !present {
		return nil
	}

	return p
}

// For Filter to generate new pack.
func (this *Engine) PipelinePack(msgLoopCount int) *PipelinePack {
	if msgLoopCount++; msgLoopCount > Globals().MaxMsgLoops {
		return nil
	}

	pack := <-this.filterRecycleChan
	pack.msgLoopCount = msgLoopCount

	return pack
}

func (this *Engine) LoadConfigFile(fn string) *Engine {
	cf, err := conf.Load(fn)
	if err != nil {
		panic(err)
	}

	this.Conf = cf

	// 'projects' section
	for i := 0; i < len(this.List("projects", nil)); i++ {
		section, err := this.Section(fmt.Sprintf("projects[%d]", i))
		if err != nil {
			panic(err)
		}

		project := &ConfProject{}
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

	return this
}

func (this *Engine) loadPluginSection(section *conf.Conf) {
	pluginCommons := new(pluginCommons)
	pluginCommons.load(section)
	if pluginCommons.disabled {
		Globals().Printf("[%s]disabled\n", pluginCommons.name)

		return
	}

	wrapper := new(pluginWrapper)
	var ok bool
	if wrapper.pluginCreator, ok = availablePlugins[pluginCommons.class]; !ok {
		pretty.Printf("allPlugins: %# v\n", availablePlugins)
		panic("unknown plugin type: " + pluginCommons.class)
	}
	wrapper.configCreator = func() *conf.Conf { return section }
	wrapper.name = pluginCommons.name

	plugin := wrapper.pluginCreator()
	plugin.Init(section)

	pluginCats := pluginTypeRegex.FindStringSubmatch(pluginCommons.class)
	if len(pluginCats) < 2 {
		panic("invalid plugin type: " + pluginCommons.class)
	}

	pluginCategory := pluginCats[1]
	if pluginCategory == "Input" {
		this.InputRunners[wrapper.name] = newInputRunner(wrapper.name, plugin.(Input),
			pluginCommons)
		this.inputWrappers[wrapper.name] = wrapper
		if pluginCommons.ticker > 0 {
			this.InputRunners[wrapper.name].setTickLength(time.Duration(pluginCommons.ticker) * time.Second)
		}

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

// ExportDiagram exports the pipeline dependencies to a diagram.
func (this *Engine) ExportDiagram(outfile string) {
	// TODO
}

func (this *Engine) ServeForever() {
	var (
		outputsWg = new(sync.WaitGroup)
		filtersWg = new(sync.WaitGroup)
		inputsWg  = new(sync.WaitGroup)

		globals = Globals()
		err     error
	)

	// setup signal handler first to avoid race condition
	// if Input terminates very soon, global.Shutdown will
	// not be able to trap it
	globals.sigChan = make(chan os.Signal)
	signal.Notify(globals.sigChan, syscall.SIGINT, syscall.SIGHUP,
		syscall.SIGUSR2, syscall.SIGUSR1)

	this.launchHttpServ()

	if globals.Verbose {
		globals.Println("Launching Output(s)...")
	}
	for _, runner := range this.OutputRunners {
		outputsWg.Add(1)
		if err = runner.start(this, outputsWg); err != nil {
			panic(err)
		}
	}

	if globals.Verbose {
		globals.Println("Launching Filter(s)...")
	}
	for _, runner := range this.FilterRunners {
		filtersWg.Add(1)
		if err = runner.start(this, filtersWg); err != nil {
			panic(err)
		}
	}

	// setup the diagnostic trackers
	inputPackTracker := newDiagnosticTracker("inputPackTracker")
	this.diagnosticTrackers[inputPackTracker.PoolName] = inputPackTracker
	filterPackTracker := newDiagnosticTracker("filterPackTracker")
	this.diagnosticTrackers[filterPackTracker.PoolName] = filterPackTracker

	if globals.Verbose {
		globals.Printf("Initializing PipelinePack pools with size=%d\n",
			globals.RecyclePoolSize)
	}
	for i := 0; i < globals.RecyclePoolSize; i++ {
		inputPack := NewPipelinePack(this.inputRecycleChan)
		inputPackTracker.AddPack(inputPack)
		this.inputRecycleChan <- inputPack

		filterPack := NewPipelinePack(this.filterRecycleChan)
		filterPackTracker.AddPack(filterPack)
		this.filterRecycleChan <- filterPack
	}

	go inputPackTracker.Run(this.Int("diagnostic_interval", 20))
	go filterPackTracker.Run(this.Int("diagnostic_interval", 20))

	// check if we have enough recycle pool reservation
	go func() {
		t := time.NewTicker(time.Second * time.Duration(globals.TickerLength))
		defer t.Stop()

		var inputPoolSize, filterPoolSize int

		for _ = range t.C {
			inputPoolSize = len(this.inputRecycleChan)
			filterPoolSize = len(this.filterRecycleChan)
			if globals.Verbose || inputPoolSize == 0 || filterPoolSize == 0 {
				globals.Printf("Recycle pool energy: [input]%d [filter]%d",
					inputPoolSize, filterPoolSize)
			}
		}
	}()

	go this.router.Start()

	for _, project := range this.projects {
		project.Start()
	}

	if globals.Verbose {
		globals.Println("Launching Input(s)...")
	}
	for _, runner := range this.InputRunners {
		inputsWg.Add(1)
		if err = runner.start(this, inputsWg); err != nil {
			inputsWg.Done()
			panic(err)
		}
	}

	for !globals.Stopping {
		select {
		case sig := <-globals.sigChan:
			globals.Printf("Got signal %s\n", sig.String())
			switch sig {
			case syscall.SIGHUP:
				globals.Println("Reloading...")
				observer.Publish(RELOAD, nil)

			case syscall.SIGINT:
				globals.Println("Engine shutdown...")
				globals.Stopping = true

			case syscall.SIGUSR1:
				observer.Publish(SIGUSR1, nil)

			case syscall.SIGUSR2:
				observer.Publish(SIGUSR2, nil)
			}
		}
	}

	// cleanup after shutdown
	inputPackTracker.Stop()
	filterPackTracker.Stop()

	this.Lock()
	for _, runner := range this.InputRunners {
		if runner == nil {
			// this Input plugin already exit
			continue
		}

		if globals.Verbose {
			globals.Printf("Stop message sent to '%s'", runner.Name())
		}

		runner.Input().Stop()
	}
	this.Unlock()
	inputsWg.Wait() // wait for all inputs done
	if globals.Verbose {
		globals.Println("All Inputs terminated")
	}

	// ok, now we are sure no more inputs, but in route.inChan there
	// still may be filter injected packs and output not consumed packs
	// we must wait for all the packs to be consumed before shutdown

	for _, runner := range this.FilterRunners {
		if globals.Verbose {
			globals.Printf("Stop message sent to '%s'", runner.Name())
		}

		this.router.removeFilterMatcher <- runner.Matcher()
	}
	filtersWg.Wait()
	if globals.Verbose {
		globals.Println("All Filters terminated")
	}

	for _, runner := range this.OutputRunners {
		if globals.Verbose {
			globals.Printf("Stop message sent to '%s'", runner.Name())
		}

		this.router.removeOutputMatcher <- runner.Matcher()
	}
	outputsWg.Wait()
	if globals.Verbose {
		globals.Println("All Outputs terminated")
	}

	//close(e.router.hub)

	this.stopHttpServ()

	for _, project := range this.projects {
		project.Stop()
	}

	globals.Printf("Shutdown with input:%s, dispatched:%s",
		gofmt.Comma(this.router.stats.TotalInputMsgN),
		gofmt.Comma(this.router.stats.TotalProcessedMsgN))
}
