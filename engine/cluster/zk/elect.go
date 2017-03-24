package zk

import (
	"github.com/funkygao/go-zookeeper/zk"
	log "github.com/funkygao/log4go"
)

func (c *controller) tryElect() {
	c.zc.SubscribeDataChanges(c.kb.controller(), c.lcl)

	c.refreshLeaderID()
	if c.leaderID != "" {
		// found leader, give up elect
		return
	}

	// elect
	if err := c.zc.CreateEphemeral(c.kb.controller(), []byte(c.participantID)); err == nil {
		// won!
		c.leaderID = c.participantID
		c.onBecomingLeader()
	} else {
		log.Trace("[%s] elect %v", c.participantID, err)
	}
}

func (c *controller) refreshLeaderID() {
	b, err := c.zc.Get(c.kb.controller())
	if err != nil && err != zk.ErrNoNode {
		log.Warn("[%s] %v", c.participantID, err)
	} else if len(b) == 0 {
		log.Warn("[%s] empty controller znode, why?", c.participantID)
	} else {
		c.leaderID = string(b)
	}
}

func (c *controller) onBecomingLeader() {
	c.zc.SubscribeChildChanges(c.kb.participants(), c.pcl)
	log.Trace("become leader!")
	c.rebalance()
}

// rebalance is called when either:
// 1. participants change
// 2. resources change
func (c *controller) rebalance() {
	participants, err := c.zc.Children(c.kb.participants())
	if err != nil {
		// TODO
		log.Error("%s", err)
		return
	}

	c.rebalanceCallback(assignResourcesToParticipants(participants, c.resources))
}
