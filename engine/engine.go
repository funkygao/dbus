package engine

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	conf "github.com/funkygao/jsconf"
	"github.com/funkygao/pretty"
	"github.com/gorilla/mux"
)

type Engine struct {
	sync.RWMutex

	*conf.Conf

	// REST exporter
	listener   net.Listener
	httpServer *http.Server
	httpRouter *mux.Router
	httpPaths  []string

	projects map[string]*ConfProject

	InputRunners  map[string]InputRunner
	inputWrappers map[string]*PluginWrapper

	FilterRunners  map[string]FilterRunner
	filterWrappers map[string]*PluginWrapper

	OutputRunners  map[string]OutputRunner
	outputWrappers map[string]*PluginWrapper

	diagnosticTrackers map[string]*DiagnosticTracker

	router *messageRouter
	stats  *EngineStats

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

	this.diagnosticTrackers = make(map[string]*DiagnosticTracker)
	this.projects = make(map[string]*ConfProject)
	this.httpPaths = make([]string, 0, 6)

	this.router = NewMessageRouter()
	this.stats = newEngineStats(this)

	this.hostname, _ = os.Hostname()
	this.pid = os.Getpid()

	return this
}

func (this *Engine) pluginNames() (names []string) {
	names = make([]string, 0, 20)
	for _, pr := range this.InputRunners {
		names = append(names, pr.Name())
	}
	for _, pr := range this.FilterRunners {
		names = append(names, pr.Name())
	}
	for _, pr := range this.OutputRunners {
		names = append(names, pr.Name())
	}

	return
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
		panic("invalid project: " + name)
	}

	return p
}

// For Filter to generate new pack
func (this *Engine) PipelinePack(msgLoopCount int) *PipelinePack {
	if msgLoopCount++; msgLoopCount > Globals().MaxMsgLoops {
		return nil
	}

	pack := <-this.filterRecycleChan
	pack.MsgLoopCount = msgLoopCount

	return pack
}

func (this *Engine) LoadConfigFile(fn string) *Engine {
	cf, err := conf.Load(fn)
	if err != nil {
		panic(err)
	}

	this.Conf = cf

	var (
		totalCpus int
		maxProcs  int
		globals   = Globals()
	)
	totalCpus = runtime.NumCPU()
	cpuNumConfig := this.String("cpu_num", "auto")
	if cpuNumConfig == "auto" {
		maxProcs = totalCpus/2 + 1
	} else {
		maxProcs, err = strconv.Atoi(cpuNumConfig)
		if err != nil {
			panic(err)
		}
	}
	runtime.GOMAXPROCS(maxProcs)

	globals.Printf("Starting engine with %d/%d CPUs...", maxProcs, totalCpus)

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

	foRunner := NewFORunner(wrapper.name, plugin, pluginCommons)
	matcher := NewMatcher(section.StringList("match", nil), foRunner)
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

// common config directives for all plugins
type pluginCommons struct {
	name     string `json:"name"`
	class    string `json:"class"`
	ticker   int    `json:"ticker_interval"`
	disabled bool   `json:"disabled"`
	comment  string `json:"comment"`
}

func (this *pluginCommons) load(section *conf.Conf) {
	this.name = section.String("name", "")
	if this.name == "" {
		pretty.Printf("%# v\n", *section)
		panic(fmt.Sprintf("invalid plugin config: %v", *section))
	}

	this.class = section.String("class", "")
	if this.class == "" {
		this.class = this.name
	}
	this.comment = section.String("comment", "")
	this.ticker = section.Int("ticker_interval", Globals().TickerLength)
	this.disabled = section.Bool("disabled", false)
}
