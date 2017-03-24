package zk

import (
	"github.com/funkygao/go-zookeeper/zk"
	log "github.com/funkygao/log4go"
)

func (c *controller) HandleNewSession() (err error) {
	log.Trace("handling new session")

	if err = c.zc.CreateLiveNode(c.kb.participant(c.participantID), nil, 3); err != nil {
		return
	}

	return nil
}

func (c *controller) HandleStateChanged(state zk.State) (err error) {
	log.Trace("zk state: %v", state)
	return
}
