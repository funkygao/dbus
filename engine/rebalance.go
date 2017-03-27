package engine

import (
	"net/http"

	log "github.com/funkygao/log4go"
)

func (e *Engine) onControllerRebalance(decision map[string][]string) {
	log.Info("decision: %+v", decision)

	for participantID, resources := range decision {
		log.Trace("[%s] rpc calling [%s]: %+v", e.participantID, participantID, resources)
		if statusCode := e.callRPC(participantID, resources); statusCode != http.StatusOK {
			log.Error("[%s] %s <- %d", e.participantID, participantID, statusCode)
			// TODO
		} else {
			log.Trace("[%s] rpc call ok [%s]: %+v", e.participantID, participantID, resources)
		}

	}
}
