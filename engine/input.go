package engine

import (
	"errors"
	"runtime/debug"
	"strings"
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

	// End notifies Input plugin it is safe to close the Acker.
	// When Input stops, Filter|Output might still depend on Input ack, that is what
	// End for.
	End(r InputRunner)

	// Run starts the main loop of the Input plugin.
	Run(r InputRunner, h PluginHelper) (err error)
}

// InputRunner is a helper for Input plugin to access some context data.
type InputRunner interface {
	PluginRunner

	// Input returns the associated Input plugin object.
	Input() Input

	// Stopper returns a channel for plugins to get notified when engine stops.
	Stopper() <-chan struct{}

	// Resources returns a channel that notifies Input plugin of the newly assigned resources in a cluster.
	// The newly assigned resource might be empty, which means the Input plugin should stop consuming the resource.
	Resources() <-chan []cluster.Resource
}

type iRunner struct {
	pRunnerBase

	inChan chan *Packet

	resourcesCh chan []cluster.Resource
	panicCh     chan<- error
}

func newInputRunner(input Input, pluginCommons *pluginCommons, panicCh chan<- error) (r *iRunner) {
	return &iRunner{
		pRunnerBase: pRunnerBase{
			plugin:        input.(Plugin),
			pluginCommons: pluginCommons,
		},
		panicCh:     panicCh,
		resourcesCh: make(chan []cluster.Resource), // FIXME how to close it
	}
}

func (ir *iRunner) Exchange() Exchange {
	return ir
}

func (ir *iRunner) Emit(pack *Packet) {
	if len(pack.Ident) == 0 {
		pack.Ident = ir.Name()
	}

	pack.acker = ir.Input()
	ir.engine.router.hub <- pack
}

func (ir *iRunner) InChan() <-chan *Packet {
	return ir.inChan
}

func (ir *iRunner) Input() Input {
	return ir.plugin.(Input)
}

func (ir *iRunner) Stopper() <-chan struct{} {
	return ir.engine.stopper
}

func (ir *iRunner) SampleConfigItems() []string {
	var r []string
	for _, line := range strings.Split(ir.plugin.SampleConfig(), "\n") {
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			r = append(r, line)
		}
	}

	return r
}

func (ir *iRunner) feedResources(resources []cluster.Resource) {
	ir.resourcesCh <- resources
}

func (ir *iRunner) Resources() <-chan []cluster.Resource {
	return ir.resourcesCh
}

func (ir *iRunner) forkAndRun(e *Engine, wg *sync.WaitGroup) {
	ir.engine = e
	ir.inChan = e.inputRecycleChans[ir.Name()]

	go ir.runMainloop(e, wg)
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
			select {
			case ir.panicCh <- reason:
			default:
				log.Warn("[%s] %s", ir.Name(), reason)
			}
		}
	}()

	globals := Globals()
	for {
		log.Trace("Input[%s] started", ir.Name())
		if err := ir.Input().Run(ir, e); err == nil {
			log.Debug("Input[%s] returned", ir.Name())
		} else {
			log.Error("Input[%s] returned: %v", ir.Name(), err)
		}

		if globals.stopping {
			return
		}

		if restart, ok := ir.plugin.(Restarter); ok {
			if !restart.CleanupForRestart() {
				// when we found all Input stopped, shutdown engine
				return
			}
		}

		log.Trace("Input[%s] restarting", ir.Name())

		// Re-initialize our plugin with its wrapper
		iw := e.inputWrappers[ir.Name()]
		ir.plugin = iw.Create()
	}
}
