package zk

import (
	"github.com/funkygao/go-zookeeper/zk"
	log "github.com/funkygao/log4go"
)

// healthCheck registers participant in zk ephemeral to allow other cluster components
// to detect failures.
// Right now our definition of health is fairly naive. If we register in zk we are healthy, otherwise
// we are dead.
type healthCheck struct {
	*controller
}

func newHealthCheck(ctx *controller) *healthCheck {
	return &healthCheck{controller: ctx}
}

func (h *healthCheck) startup() {
	h.zc.SubscribeStateChanges(h)
	h.register()
}

func (h *healthCheck) register() (err error) {
	if err = h.zc.CreateLiveNode(h.kb.participant(h.participantID), nil, 2); err != nil {
		return
	}

	return
}

func (h *healthCheck) HandleNewSession() (err error) {
	log.Trace("[%s] bring participant alive", h.participantID)
	h.register()
	return
}

func (h *healthCheck) HandleStateChanged(state zk.State) (err error) {
	log.Trace("[%s] %s", h.participantID, state)
	return
}
