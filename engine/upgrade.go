package engine

import (
	"net/http"
	"os"

	log "github.com/funkygao/log4go"
	"github.com/inconshreveable/go-update"
)

func (e *Engine) watchUpgrade(evts <-chan struct{}) {
	endpoint := os.Getenv("UPGRADE_ENDPOINT")
	if len(endpoint) == 0 {
		evts = nil
		log.Warn("upgrade events ignored")
	}

	for {
		select {
		case <-evts:
			log.Info("upgrading...")
			if err := e.doUpgrade(endpoint); err != nil {
				log.Error("upgrade: %v", err)
			}

		case <-e.stopper:
			return
		}
	}

}

func (e *Engine) doUpgrade(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return update.Apply(resp.Body, update.Options{})
}
