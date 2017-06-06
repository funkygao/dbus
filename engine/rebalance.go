package engine

import (
	"fmt"
	"net/http"

	"github.com/funkygao/dbus"
	"github.com/funkygao/dbus/pkg/cluster"
	"github.com/funkygao/gorequest"
	log "github.com/funkygao/log4go"
)

func (e *Engine) leaderRebalance(epoch int, decision cluster.Decision) {
	log.Info("[%s] decision: %+v", e.participant, decision)

	// 2 phase commit
	for phase := 1; phase <= 2; phase++ {
		for participant, resources := range decision {
			log.Debug("[%s-%d] rpc-> %s %+v", e.participant, phase, participant.Endpoint, resources)

			// edge case:
			// participant might die
			// participant not die, but its Input plugin panic
			// leader might die
			// rpc might be rejected
			statusCode := e.callRPC(participant.Endpoint, epoch, phase, resources)
			switch statusCode {
			case http.StatusOK:
				log.Trace("[%s-%d] rpc<- ok %s", e.participant, phase, participant.Endpoint)

			case http.StatusGone:
				// e,g.
				// resource changed, live participant [1, 2, 3], when RPC sending, p[1] gone
				// just wait for another rebalance event
				log.Warn("[%s-%d] rpc<- %s gone", e.participant, phase, participant.Endpoint)
				return

			case http.StatusBadRequest:
				// should never happen
				log.Critical("[%s-%d] rpc<- %s bad request", e.participant, phase, participant.Endpoint)
				e.callSOS("[%s-%d] rpc<- %s bad request", e.participant, phase, participant.Endpoint)

			case http.StatusNotAcceptable:
				log.Error("[%s-%d] rpc<- %s leader moved, await next rebalance", e.participant, phase, participant.Endpoint)
				return

			default:
				log.Error("[%s-%d] rpc<- %s %d, trigger new rebalance!", e.participant, phase, participant.Endpoint, statusCode)
				e.callSOS("[%s-%d] rpc<- %s %d, trigger new rebalance!", e.participant, phase, participant.Endpoint, statusCode)

				if err := e.ClusterManager().Rebalance(); err != nil {
					log.Critical("%s", err)
				}
			}

		}
	}
}

func (e *Engine) callRPC(endpoint string, epoch int, phase int, resources []cluster.Resource) int {
	resp, _, errs := gorequest.New().
		Post(fmt.Sprintf("http://%s/v1/rebalance?epoch=%d&phase=%d", endpoint, epoch, phase)).
		Set("User-Agent", fmt.Sprintf("dbus-%s", dbus.Revision)).
		SendString(string(cluster.Resources(resources).Marshal())).
		End()
	if len(errs) > 0 {
		// e,g. participant gone
		// connection reset
		// connnection refused
		// FIXME what if connection timeout?
		return http.StatusGone
	}

	return resp.StatusCode
}
