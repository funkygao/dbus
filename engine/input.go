package engine

import (
	"runtime/debug"
	"sync"

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
}

type iRunner struct {
	pRunnerBase

	inChan chan *Packet
}

func newInputRunner(input Input, pluginCommons *pluginCommons) (r InputRunner) {
	return &iRunner{
		pRunnerBase: pRunnerBase{
			plugin:        input.(Plugin),
			pluginCommons: pluginCommons,
		},
	}
}

func (ir *iRunner) Inject(pack *Packet) {
	if pack.Ident == "" {
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

func (ir *iRunner) start(e *Engine, wg *sync.WaitGroup) error {
	ir.engine = e
	ir.inChan = e.inputRecycleChan

	go ir.runMainloop(e, wg)
	return nil
}

func (ir *iRunner) runMainloop(e *Engine, wg *sync.WaitGroup) {
	defer func() {
		if err := recover(); err != nil {
			log.Critical("[%s] %v\n%s", ir.Name(), err, string(debug.Stack()))
		}

		wg.Done()
	}()

	globals := Globals()
	for {
		log.Info("Input[%s] starting", ir.Name())
		if err := ir.Input().Run(ir, e); err == nil {
			log.Info("Input[%s] stopped", ir.Name())
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
