package engine

import (
	"time"

	log "github.com/funkygao/log4go"
)

// diagnosticTracker is a diagnostic tracker for the pipeline packs pool.
type diagnosticTracker struct {
	PoolName string

	// All the packs in a recycle pool
	packs []*PipelinePack

	stopChan chan interface{}
}

func newDiagnosticTracker(poolName string) *diagnosticTracker {
	return &diagnosticTracker{packs: make([]*PipelinePack, 0, Globals().RecyclePoolSize),
		PoolName: poolName, stopChan: make(chan interface{})}
}

func (this *diagnosticTracker) AddPack(pack *PipelinePack) {
	this.packs = append(this.packs, pack)
}

func (this *diagnosticTracker) Run(interval int) {
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

	log.Trace("Diagnostic[%s] started with %ds", this.PoolName, interval)

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
				log.Warn("[%s]%d packs have been idle more than %.0f seconds",
					this.PoolName, len(probablePacks), idleMax.Seconds())
				for runner, count = range pluginCounts {
					runner.setLeakCount(count) // let runner know leak count

					log.Warn("\t%s: %d", runner.Name(), count)
				}

				for _, pack := range probablePacks {
					log.Debug("[%s]%s", this.PoolName, *pack)
				}
			}

		case <-this.stopChan:
			ever = false
		}
	}
}

func (this *diagnosticTracker) Stop() {
	log.Trace("Diagnostic[%s] stopped", this.PoolName)

	close(this.stopChan)
}
