package engine

import (
	"sync"

	log "github.com/funkygao/log4go"
)

var (
	_ PluginRunner = &iRunner{}
	_ InputRunner  = &iRunner{}
)

type Input interface {
	Plugin

	Run(r InputRunner, h PluginHelper) (err error)
	Stop(InputRunner)
}

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

func newInputRunner(name string, input Input, pluginCommons *pluginCommons) (r InputRunner) {
	return &iRunner{
		pRunnerBase: pRunnerBase{
			name:          name,
			plugin:        input.(Plugin),
			pluginCommons: pluginCommons,
		},
	}
}

func (ir *iRunner) Inject(pack *Packet) {
	if pack.Ident == "" {
		pack.Ident = ir.name
	}
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
			log.Critical("[%s] %v", ir.name, err)
		}

		wg.Done()
	}()

	globals := Globals()
	for {
		log.Info("Input[%s] starting", ir.name)
		if err := ir.Input().Run(ir, e); err == nil {
			log.Info("Input[%s] stopped", ir.name)
		} else {
			log.Error("Input[%s] stopped: %v", ir.name, err)
		}

		if globals.Stopping {
			e.stopInputRunner(ir.name)

			return
		}

		if restart, ok := ir.plugin.(Restarting); ok {
			if !restart.CleanupForRestart() {
				// when we found all Input stopped, shutdown engine
				e.stopInputRunner(ir.name)

				return
			}
		}

		log.Trace("Input[%s] restarting", ir.name)

		// Re-initialize our plugin with its wrapper
		iw := e.inputWrappers[ir.name]
		ir.plugin = iw.Create()
	}

}
