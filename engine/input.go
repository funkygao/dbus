package engine

import (
	"errors"
	"runtime/debug"
	"sync"

	"github.com/funkygao/dbus/pkg/cluster"
	log "github.com/funkygao/log4go"
)

var (
	_ PluginRunner = &iRunner{}
	_ InputRunner  = &iRunner{}
)

// Input is the input plugin.
type Input interface {
	Plugin

	Acker

	// Run starts the main loop of the Input plugin.
	Run(r InputRunner, h PluginHelper) (err error)

	// Stop is the callback which stops the Input plugin.
	Stop(InputRunner)
}

// InputRunner is a helper for Input plugin to access some context data.
type InputRunner interface {
	PluginRunner

	// InChan returns input channel from which Inputs can get fresh Packets.
	InChan() chan *Packet

	// Input returns the associated Input plugin object.
	Input() Input

	// Injects Packet into the Router's input channel for delivery
	// to all Filter and Output plugins with corresponding matcher.
	Inject(pack *Packet)

	// Resources returns a channel that notifies Input plugin of the newly assigned resources in a cluster.
	// The newly assigned resource might be empty, which means the Input plugin should stop consuming the resource.
	Resources() <-chan []cluster.Resource
}

type iRunner struct {
	pRunnerBase

	inChan chan *Packet

	resourcesCh chan []cluster.Resource
	panicCh     chan error
}

func newInputRunner(input Input, pluginCommons *pluginCommons, panicCh chan error) (r *iRunner) {
	return &iRunner{
		pRunnerBase: pRunnerBase{
			plugin:        input.(Plugin),
			pluginCommons: pluginCommons,
		},
		panicCh:     panicCh,
		resourcesCh: make(chan []cluster.Resource), // FIXME how to close it
	}
}

func (ir *iRunner) Inject(pack *Packet) {
	if len(pack.Ident) == 0 {
		pack.Ident = ir.Name()
	}

	pack.input = ir.Input()
	ir.engine.router.hub <- pack
}

func (ir *iRunner) InChan() chan *Packet {
	return ir.inChan
}

func (ir *iRunner) Input() Input {
	return ir.plugin.(Input)
}

func (ir *iRunner) feedResources(resources []cluster.Resource) {
	ir.resourcesCh <- resources
}

func (ir *iRunner) Resources() <-chan []cluster.Resource {
	return ir.resourcesCh
}

func (ir *iRunner) start(e *Engine, wg *sync.WaitGroup) error {
	ir.engine = e
	ir.inChan = e.inputRecycleChans[ir.Name()]

	go ir.runMainloop(e, wg)
	return nil
}

func (ir *iRunner) runMainloop(e *Engine, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()

		if err := recover(); err != nil {
			log.Critical("[%s] shutdown completely for: %v\n%s", ir.Name(), err, string(debug.Stack()))

			reason := errors.New("unexpected reason")
			switch panicErr := err.(type) {
			case string:
				reason = errors.New(panicErr)
			case error:
				reason = panicErr
			}
			ir.panicCh <- reason
		}
	}()

	globals := Globals()
	for {
		log.Trace("Input[%s] started", ir.Name())
		if err := ir.Input().Run(ir, e); err == nil {
			log.Trace("Input[%s] stopped", ir.Name())
		} else {
			log.Error("Input[%s] stopped: %v", ir.Name(), err)
		}

		if globals.Stopping {
			e.stopInputRunner(ir.Name())

			return
		}

		if restart, ok := ir.plugin.(Restarter); ok {
			if !restart.CleanupForRestart() {
				// when we found all Input stopped, shutdown engine
				e.stopInputRunner(ir.Name())

				return
			}
		}

		log.Trace("Input[%s] restarting", ir.Name())

		// Re-initialize our plugin with its wrapper
		iw := e.inputWrappers[ir.Name()]
		ir.plugin = iw.Create()
	}
}
