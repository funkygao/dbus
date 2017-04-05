package engine

import (
	"os"

	log "github.com/funkygao/log4go"
)

func (e *Engine) watchUpgrade(evts <-chan struct{}) {
	endpoint := os.Getenv("UPGRADE_ENDPOINT")
	if len(endpoint) == 0 {

	}

	for {
		select {
		case <-evts:
			log.Info("upgrading...")

		case <-e.stopper:
			return
		}
	}

}
