package engine

import (
	"os"
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

		case <-e.stopper:
			return
		}
	}
}

func (e *Engine) watchConfig(notifier chan struct{}, poller time.Duration) {
	fn := e.Conf.ConfPath().S()
	log.Info("watching %s with poller=%s", fn, poller)

	lastStat, err := os.Stat(fn)
	if err != nil {
		panic(err)
	}

	ticker := time.NewTicker(poller)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stat, err := os.Stat(fn)
			if err != nil {
				panic(err)
			}

			if stat.ModTime() != lastStat.ModTime() {
				close(notifier)
				return
			}

		case <-e.stopper:
			return
		}
	}
}
