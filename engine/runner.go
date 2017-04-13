package engine

import (
	"errors"
	"runtime/debug"
	"sync"

	conf "github.com/funkygao/jsconf"
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

	// Exchange returns an Exchange for the plugin to exchange packets.
	Exchange() Exchange

	// Plugin returns the underlying plugin object.
	Plugin() Plugin

	// Conf returns the underlying plugin specific configuration.
	Conf() *conf.Conf

	start(e *Engine, wg *sync.WaitGroup, stopper <-chan struct{}) (err error)
}

// FilterOutputRunner is the common interface shared by FilterRunner and OutputRunner.
type FilterOutputRunner interface {
	PluginRunner

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

func (pb *pRunnerBase) Conf() *conf.Conf {
	return pb.pluginCommons.cf
}

// foRunner is filter/output runner.
type foRunner struct {
	pRunnerBase

	matcher *matcher

	inChan  chan *Packet
	panicCh chan<- error
}

func newFORunner(plugin Plugin, pluginCommons *pluginCommons, panicCh chan<- error) *foRunner {
	return &foRunner{
		pRunnerBase: pRunnerBase{
			plugin:        plugin,
			pluginCommons: pluginCommons,
		},
		inChan:  make(chan *Packet, Globals().PluginChanSize),
		panicCh: panicCh,
	}
}

func (fo *foRunner) getMatcher() *matcher {
	return fo.matcher
}

func (fo *foRunner) Ack(pack *Packet) error {
	return pack.ack()
}

func (fo *foRunner) Inject(pack *Packet) {
	fo.engine.router.hub <- pack
}

func (fo *foRunner) Exchange() Exchange {
	return fo
}

func (fo *foRunner) InChan() <-chan *Packet {
	return fo.inChan
}

func (fo *foRunner) Output() Output {
	return fo.plugin.(Output)
}

func (fo *foRunner) Filter() Filter {
	return fo.plugin.(Filter)
}

func (fo *foRunner) start(e *Engine, wg *sync.WaitGroup, stopper <-chan struct{}) error {
	fo.engine = e

	go fo.runMainloop(wg, stopper)
	return nil
}

func (fo *foRunner) runMainloop(wg *sync.WaitGroup, stopper <-chan struct{}) {
	defer func() {
		wg.Done()

		if err := recover(); err != nil {
			log.Critical("[%s] shutdown completely for: %v\n%s", fo.Name(), err, string(debug.Stack()))

			reason := errors.New("unexpected reason")
			switch panicErr := err.(type) {
			case string:
				reason = errors.New(panicErr)
			case error:
				reason = panicErr
			}
			fo.panicCh <- reason
		}
	}()

	var (
		pluginType string
		pw         *pluginWrapper
	)

	globals := Globals()
	for {
		if filter, ok := fo.plugin.(Filter); ok {
			log.Info("Filter[%s] started", fo.Name())

			pluginType = "filter"
			if err := filter.Run(fo, fo.engine, stopper); err != nil {
				log.Error("Filter[%s] stopped: %v", fo.Name(), err)
			} else {
				log.Info("Filter[%s] stopped", fo.Name())
			}
		} else if output, ok := fo.plugin.(Output); ok {
			log.Info("Output[%s] started", fo.Name())

			pluginType = "output"
			if err := output.Run(fo, fo.engine, stopper); err != nil {
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
