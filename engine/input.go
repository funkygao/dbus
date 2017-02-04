package engine

import (
	"sync"
)

var (
	_ PluginRunner = &iRunner{}
	_ InputRunner  = &iRunner{}
)

type Input interface {
	Plugin

	Run(r InputRunner, h PluginHelper) (err error)
	Stop()
}

type InputRunner interface {
	PluginRunner

	// InChan returns input channel from which Inputs can get fresh PipelinePacks.
	InChan() chan *PipelinePack

	// Input returns the associated Input plugin object.
	Input() Input

	// Injects PipelinePack into the Router's input channel for delivery
	// to all Filter and Output plugins with corresponding matcher.
	Inject(pack *PipelinePack)
}

type iRunner struct {
	pRunnerBase

	ident  string
	inChan chan *PipelinePack
}

func newInputRunner(name string, input Input, pluginCommons *pluginCommons, ident string) (r InputRunner) {
	return &iRunner{
		pRunnerBase: pRunnerBase{
			name:          name,
			plugin:        input.(Plugin),
			pluginCommons: pluginCommons,
		},
		ident: ident,
	}
}

func (this *iRunner) Inject(pack *PipelinePack) {
	pack.input = true
	if pack.Ident == "" {
		pack.Ident = this.ident
	}
	this.engine.router.hub <- pack
}

func (this *iRunner) InChan() chan *PipelinePack {
	return this.inChan
}

func (this *iRunner) Input() Input {
	return this.plugin.(Input)
}

func (this *iRunner) start(e *Engine, wg *sync.WaitGroup) error {
	this.engine = e

	// got the engine's recylable PipelinePack pool
	this.inChan = e.inputRecycleChan

	go this.runMainloop(e, wg)
	return nil
}

func (this *iRunner) runMainloop(e *Engine, wg *sync.WaitGroup) {
	defer wg.Done()

	globals := Globals()
	for {
		if globals.Verbose {
			globals.Printf("Input[%s] starting", this.name)
		}

		if err := this.Input().Run(this, e); err != nil {
			panic(err)
		}

		if globals.Verbose {
			globals.Printf("Input[%s] ended", this.name)
		}

		if globals.Stopping {
			e.stopInputRunner(this.name)

			return
		}

		if restart, ok := this.plugin.(Restarting); ok {
			if !restart.CleanupForRestart() {
				// when we found all Input stopped, shutdown engine
				e.stopInputRunner(this.name)

				return
			}
		}

		if globals.Verbose {
			globals.Printf("Input[%s] restarting", this.name)
		}

		// Re-initialize our plugin with its wrapper
		iw := e.inputWrappers[this.name]
		this.plugin = iw.Create()
	}

}
