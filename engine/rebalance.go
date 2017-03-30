package engine

import (
	"fmt"
	"net/http"

	"github.com/funkygao/dbus"
	"github.com/funkygao/dbus/pkg/cluster"
	"github.com/funkygao/gorequest"
	log "github.com/funkygao/log4go"
)

func (e *Engine) onControllerRebalance(decision cluster.Decision) {
	log.Info("decision: %+v", decision)

	for participant, resources := range decision {
		log.Trace("[%s] rpc-> %s %+v", e.participant, participant.Endpoint, resources)

		if statusCode := e.callRPC(participant.Endpoint, resources); statusCode != http.StatusOK {
			log.Error("[%s] rpc<- %s %d", e.participant, participant.Endpoint, statusCode)
			// TODO
		} else {
			log.Trace("[%s] rpc<- ok %s", e.participant, participant.Endpoint)
		}
	}
}

func (e *Engine) callRPC(endpoint string, resources []cluster.Resource) int {
	resp, _, err := gorequest.New().
		Post(fmt.Sprintf("http://%s/v1/rebalance", endpoint)).
		Set("User-Agent", fmt.Sprintf("dbus-%s", dbus.Revision)).
		SendString(string(cluster.Resources(resources).Marshal())).
		End()
	if err != nil {
		return http.StatusBadRequest
	}

	return resp.StatusCode
}
