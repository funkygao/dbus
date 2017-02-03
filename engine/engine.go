package engine

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	conf "github.com/funkygao/jsconf"
	"github.com/funkygao/pretty"
	"github.com/gorilla/mux"
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
	inputWrappers map[string]*PluginWrapper

	FilterRunners  map[string]FilterRunner
	filterWrappers map[string]*PluginWrapper

	OutputRunners  map[string]OutputRunner
	outputWrappers map[string]*PluginWrapper

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

func New(globals *GlobalConfigStruct) (this *Engine) {
	this = new(Engine)

	if globals == nil {
		globals = DefaultGlobals()
	}
	Globals = func() *GlobalConfigStruct {
		return globals
	}

	this.InputRunners = make(map[string]InputRunner)
	this.inputWrappers = make(map[string]*PluginWrapper)
	this.FilterRunners = make(map[string]FilterRunner)
	this.filterWrappers = make(map[string]*PluginWrapper)
	this.OutputRunners = make(map[string]OutputRunner)
	this.outputWrappers = make(map[string]*PluginWrapper)

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

	wrapper := new(PluginWrapper)
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
		this.InputRunners[wrapper.name] = NewInputRunner(wrapper.name, plugin.(Input),
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
func (tihs *Engine) ExportDiagram(outfile string) {
	// TODO
}
