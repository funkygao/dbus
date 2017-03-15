package engine

import (
	"time"

	log "github.com/funkygao/log4go"
)

func (e *Engine) runWatchdog(interval time.Duration) {
	// not a goroutine leakage because it will run once even when engine reload
	for {
		inputPoolSize := len(e.inputRecycleChan)
		filterPoolSize := len(e.filterRecycleChan)
		if inputPoolSize == 0 || filterPoolSize == 0 {
			log.Warn("Recycle pool reservation: [input]%d [filter]%d", inputPoolSize, filterPoolSize)
		}

		time.Sleep(interval)
	}
}
