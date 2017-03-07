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
	name          string
	plugin        Plugin
	engine        *Engine
	pluginCommons *pluginCommons
}

func (pb *pRunnerBase) Name() string {
	return pb.name
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

func newFORunner(name string, plugin Plugin, pluginCommons *pluginCommons) *foRunner {
	return &foRunner{
		pRunnerBase: pRunnerBase{
			name:          name,
			plugin:        plugin,
			pluginCommons: pluginCommons,
		},
		inChan: make(chan *Packet, Globals().PluginChanSize),
	}
}

func (fo *foRunner) getMatcher() *matcher {
	return fo.matcher
}

func (fo *foRunner) Inject(pack *Packet) bool {
	pack.input = false
	fo.engine.router.hub <- pack
	return true
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
			log.Critical("[%s] %v", fo.name, err)
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
			log.Trace("Filter[%s] starting", fo.name)

			pluginType = "filter"
			if err := filter.Run(fo, fo.engine); err != nil {
				log.Trace("Filter[%s] stopped: %v", fo.name, err)
			} else {
				log.Trace("Filter[%s] stopped", fo.name)
			}
		} else if output, ok := fo.plugin.(Output); ok {
			log.Trace("Output[%s] starting", fo.name)

			pluginType = "output"
			if err := output.Run(fo, fo.engine); err != nil {
				log.Trace("Output[%s] stopped: %v", fo.name, err)
			} else {
				log.Trace("Output[%s] stopped", fo.name)
			}
		} else {
			panic("unknown plugin type")
		}

		if globals.Stopping {
			return
		}

		if restart, ok := fo.plugin.(Restarting); ok {
			if !restart.CleanupForRestart() {
				return
			}
		}

		log.Trace("[%s] restarting", fo.name)

		// Re-initialize our plugin using its wrapper
		if pluginType == "filter" {
			pw = fo.engine.filterWrappers[fo.name]
		} else {
			pw = fo.engine.outputWrappers[fo.name]
		}
		fo.plugin = pw.Create()
	}

}
