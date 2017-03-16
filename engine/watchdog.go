package engine

import (
	"time"

	log "github.com/funkygao/log4go"
)

func (e *Engine) runWatchdog(interval time.Duration) {
	tick := time.NewTicker(interval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			inputPoolSize := len(e.inputRecycleChan)
			filterPoolSize := len(e.filterRecycleChan)
			if inputPoolSize == 0 || filterPoolSize == 0 {
				log.Warn("Recycle pool reservation: [input]%d [filter]%d", inputPoolSize, filterPoolSize)
			}

		case <-e.stopper:
			return
		}
	}
}
