package engine

import (
	"sync"
)

// PluginRunner is the base interface for the  plugin runners.
type PluginRunner interface {

	// Name returns the name of the underlying plugin.
	Name() string

	// Underlying plugin object.
	Plugin() Plugin

	start(e *Engine, wg *sync.WaitGroup) (err error)

	setLeakCount(count int)
	LeakCount() int
}

// Filter and Output runner extends PluginRunner
type FilterOutputRunner interface {
	PluginRunner

	InChan() chan *PipelinePack
	Matcher() *matcher
}

// Base for all runners
type pRunnerBase struct {
	name          string
	plugin        Plugin
	engine        *Engine
	pluginCommons *pluginCommons
	leakCount     int
}

// foRunner is filter output runner.
type foRunner struct {
	pRunnerBase

	matcher   *matcher
	inChan    chan *PipelinePack
	leakCount int
}

func (this *pRunnerBase) Name() string {
	return this.name
}

func (this *pRunnerBase) Plugin() Plugin {
	return this.plugin
}

func (this *pRunnerBase) setLeakCount(count int) {
	this.leakCount = count
}

func (this *pRunnerBase) LeakCount() int {
	return this.leakCount
}

func newFORunner(name string, plugin Plugin, pluginCommons *pluginCommons) (this *foRunner) {
	this = &foRunner{
		pRunnerBase: pRunnerBase{
			name:          name,
			plugin:        plugin,
			pluginCommons: pluginCommons,
		},
		inChan: make(chan *PipelinePack, Globals().PluginChanSize),
	}

	return
}

func (this *foRunner) Matcher() *matcher {
	return this.matcher
}

func (this *foRunner) Inject(pack *PipelinePack) bool {
	pack.input = false
	this.engine.router.hub <- pack
	return true
}

func (this *foRunner) InChan() chan *PipelinePack {
	return this.inChan
}

func (this *foRunner) Output() Output {
	return this.plugin.(Output)
}

func (this *foRunner) Filter() Filter {
	return this.plugin.(Filter)
}

func (this *foRunner) start(e *Engine, wg *sync.WaitGroup) error {
	this.engine = e

	go this.runMainloop(wg)
	return nil
}

func (this *foRunner) runMainloop(wg *sync.WaitGroup) {
	defer wg.Done()

	var (
		pluginType string
		pw         *pluginWrapper
	)

	globals := Globals()
	for {
		if filter, ok := this.plugin.(Filter); ok {
			if globals.Verbose {
				globals.Printf("Filter[%s] starting", this.name)
			}

			pluginType = "filter"
			filter.Run(this, this.engine)

			if globals.Verbose {
				globals.Printf("Filter[%s] ended", this.name)
			}
		} else if output, ok := this.plugin.(Output); ok {
			if globals.Verbose {
				globals.Printf("Output[%s] starting", this.name)
			}

			pluginType = "output"
			output.Run(this, this.engine)

			if globals.Verbose {
				globals.Printf("Output[%s] ended", this.name)
			}
		} else {
			panic("unknown plugin type")
		}

		if globals.Stopping {
			return
		}

		if restart, ok := this.plugin.(Restarting); ok {
			if !restart.CleanupForRestart() {
				return
			}
		}

		if globals.Verbose {
			globals.Printf("[%s] restarting", this.name)
		}

		// Re-initialize our plugin using its wrapper
		if pluginType == "filter" {
			pw = this.engine.filterWrappers[this.name]
		} else {
			pw = this.engine.outputWrappers[this.name]
		}
		this.plugin = pw.Create()
	}

}
