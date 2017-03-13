package engine

import (
	"sync"

	log "github.com/funkygao/log4go"
)

var (
	_ PluginRunner       = &foRunner{}
	_ FilterOutputRunner = &foRunner{}
	_ OutputRunner       = &foRunner{}
	_ FilterRunner       = &foRunner{}
)

// PluginRunner is the base interface for the plugin runners.
type PluginRunner interface {

	// Name returns the name of the underlying plugin.
	Name() string

	// Class returns the class name of the underlying plugin.
	Class() string

	// Plugin returns the underlying plugin object.
	Plugin() Plugin

	start(e *Engine, wg *sync.WaitGroup) (err error)
}

// Filter and Output runner extends PluginRunner
type FilterOutputRunner interface {
	PluginRunner

	InChan() chan *Packet

	getMatcher() *matcher
}

// pRunnerBase is base for all plugin runners.
type pRunnerBase struct {
	plugin        Plugin
	engine        *Engine
	pluginCommons *pluginCommons
}

func (pb *pRunnerBase) Name() string {
	return pb.pluginCommons.name
}

func (pb *pRunnerBase) Class() string {
	return pb.pluginCommons.class
}

func (pb *pRunnerBase) Plugin() Plugin {
	return pb.plugin
}

// foRunner is filter output runner.
type foRunner struct {
	pRunnerBase

	matcher   *matcher
	inChan    chan *Packet
	leakCount int
}

func newFORunner(plugin Plugin, pluginCommons *pluginCommons) *foRunner {
	return &foRunner{
		pRunnerBase: pRunnerBase{
			plugin:        plugin,
			pluginCommons: pluginCommons,
		},
		inChan: make(chan *Packet, Globals().PluginChanSize),
	}
}

func (fo *foRunner) getMatcher() *matcher {
	return fo.matcher
}

func (fo *foRunner) Inject(pack *Packet) {
	fo.engine.router.hub <- pack
}

func (fo *foRunner) InChan() chan *Packet {
	return fo.inChan
}

func (fo *foRunner) Output() Output {
	return fo.plugin.(Output)
}

func (fo *foRunner) Filter() Filter {
	return fo.plugin.(Filter)
}

func (fo *foRunner) start(e *Engine, wg *sync.WaitGroup) error {
	fo.engine = e

	go fo.runMainloop(wg)
	return nil
}

func (fo *foRunner) runMainloop(wg *sync.WaitGroup) {
	defer func() {
		if err := recover(); err != nil {
			log.Critical("[%s] %v", fo.Name(), err)
		}

		wg.Done()
	}()

	var (
		pluginType string
		pw         *pluginWrapper
	)

	globals := Globals()
	for {
		if filter, ok := fo.plugin.(Filter); ok {
			log.Info("Filter[%s] starting", fo.Name())

			pluginType = "filter"
			if err := filter.Run(fo, fo.engine); err != nil {
				log.Error("Filter[%s] stopped: %v", fo.Name(), err)
			} else {
				log.Info("Filter[%s] stopped", fo.Name())
			}
		} else if output, ok := fo.plugin.(Output); ok {
			log.Info("Output[%s] starting", fo.Name())

			pluginType = "output"
			if err := output.Run(fo, fo.engine); err != nil {
				log.Error("Output[%s] stopped: %v", fo.Name(), err)
			} else {
				log.Info("Output[%s] stopped", fo.Name())
			}
		} else {
			panic("unknown plugin type")
		}

		if globals.Stopping {
			return
		}

		if restart, ok := fo.plugin.(Restarter); ok {
			if !restart.CleanupForRestart() {
				return
			}
		}

		log.Trace("[%s] restarting", fo.Name())

		// Re-initialize our plugin using its wrapper
		if pluginType == "filter" {
			pw = fo.engine.filterWrappers[fo.Name()]
		} else {
			pw = fo.engine.outputWrappers[fo.Name()]
		}
		fo.plugin = pw.Create()
	}

}
