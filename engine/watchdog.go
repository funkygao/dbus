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
