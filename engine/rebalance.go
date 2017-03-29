package engine

import (
	"net/http"

	"github.com/funkygao/dbus/pkg/cluster"
	log "github.com/funkygao/log4go"
)

func (e *Engine) onControllerRebalance(decision cluster.Decision) {
	log.Info("decision: %+v", decision)

	for participant, resources := range decision {
		log.Trace("[%s] rpc calling [%s]: %+v", e.participant, participant.Endpoint, resources)

		if statusCode := e.callRPC(participant.Endpoint, resources); statusCode != http.StatusOK {
			log.Error("[%s] %s <- %d", e.participant, participant.Endpoint, statusCode)
			// TODO
		} else {
			log.Trace("[%s] rpc call ok [%s]", e.participant, participant.Endpoint)
		}

	}
}
