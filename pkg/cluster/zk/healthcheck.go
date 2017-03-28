package zk

import (
	"github.com/funkygao/go-zookeeper/zk"
	log "github.com/funkygao/log4go"
	"github.com/funkygao/zkclient"
)

var (
	_ zkclient.ZkStateListener = &healthCheck{}
)

// healthCheck registers participant in zk ephemeral to allow other cluster components
// to detect failures.
// Right now our definition of health is fairly naive. If we register in zk we are healthy, otherwise
// we are dead.
type healthCheck struct {
	participantID string
	*zkclient.Client
	*keyBuilder
}

func newHealthCheck(participantID string, zc *zkclient.Client, kb *keyBuilder) *healthCheck {
	return &healthCheck{Client: zc, keyBuilder: kb, participantID: participantID}
}

func (h *healthCheck) startup() {
	h.SubscribeStateChanges(h)
	h.register()
}

func (h *healthCheck) register() {
	if err := h.CreateLiveNode(h.participant(h.participantID), nil, 2); err != nil {
		// 2 same participant running?
		panic(err)
	}

	log.Trace("[%s] alive", h.participantID)
}

func (h *healthCheck) HandleNewSession() (err error) {
	h.register()
	return
}

func (h *healthCheck) HandleStateChanged(state zk.State) (err error) {
	log.Trace("[%s] %s", h.participantID, state)
	return
}
