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
			inputChanFull := false
			inputs := make(map[string]int, len(e.inputRecycleChans))
			for name, ch := range e.inputRecycleChans {
				inputs[name] = len(ch)
				if inputs[name] == 0 {
					inputChanFull = true
				}
			}
			filterPoolSize := len(e.filterRecycleChan)
			if inputChanFull || filterPoolSize == 0 {
				log.Warn("Recycle pool reservation: [filter]%d, inputs %v", filterPoolSize, inputs)
			}

		case <-e.upgradeStopper:
			return
		}
	}
}
