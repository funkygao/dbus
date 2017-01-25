package engine

import (
	"time"
)

// A diagnostic tracker for the pipeline packs pool
type DiagnosticTracker struct {
	PoolName string

	// All the packs in a recycle pool
	packs []*PipelinePack

	stopChan chan interface{}
}

func NewDiagnosticTracker(poolName string) *DiagnosticTracker {
	return &DiagnosticTracker{packs: make([]*PipelinePack, 0, Globals().RecyclePoolSize),
		PoolName: poolName, stopChan: make(chan interface{})}
}

func (this *DiagnosticTracker) AddPack(pack *PipelinePack) {
	this.packs = append(this.packs, pack)
}

func (this *DiagnosticTracker) Run(interval int) {
	var (
		pack           *PipelinePack
		earliestAccess time.Time
		pluginCounts   map[PluginRunner]int
		count          int
		runner         PluginRunner
		globals        = Globals()
		ever           = true
	)

	idleMax := globals.MaxPackIdle
	probablePacks := make([]*PipelinePack, 0, len(this.packs))
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	if globals.Debug {
		globals.Printf("Diagnostic[%s] started with %ds",
			this.PoolName, interval)
	}

	for ever {
		select {
		case <-ticker.C:

			probablePacks = probablePacks[:0] // reset
			pluginCounts = make(map[PluginRunner]int)

			// Locate all the packs that have not been touched in idleMax duration
			// that are not recycled
			earliestAccess = time.Now().Add(-idleMax)
			for _, pack = range this.packs {
				if pack.diagnostics.RunnerCount() == 0 {
					continue
				}

				if pack.diagnostics.LastAccess.Before(earliestAccess) {
					probablePacks = append(probablePacks, pack)
					for _, runner = range pack.diagnostics.Runners() {
						pluginCounts[runner] += 1
					}
				}
			}

			if len(probablePacks) > 0 {
				globals.Printf("[%s]%d packs have been idle more than %.0f seconds",
					this.PoolName, len(probablePacks), idleMax.Seconds())
				for runner, count = range pluginCounts {
					runner.setLeakCount(count) // let runner know leak count

					globals.Printf("\t%s: %d", runner.Name(), count)
				}

				if globals.Debug {
					for _, pack := range probablePacks {
						globals.Printf("[%s]%s", this.PoolName, *pack)
					}
				}
			}

		case <-this.stopChan:
			ever = false
		}
	}
}

func (this *DiagnosticTracker) Stop() {
	globals := Globals()
	if globals.Verbose {
		globals.Printf("Diagnostic[%s] stopped", this.PoolName)
	}

	close(this.stopChan)
}
