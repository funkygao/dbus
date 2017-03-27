package zk

import (
	"github.com/funkygao/go-zookeeper/zk"
	log "github.com/funkygao/log4go"
)

func (c *controller) tryElect() {
	log.Trace("[%s] try electing...", c.participantID)

	c.zc.SubscribeDataChanges(c.kb.controller(), c.lcl)

	c.refreshLeaderID()
	if c.leaderID != "" {
		log.Trace("[%s] found leader: %s", c.participantID, c.leaderID)
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
	} else if err == nil && len(b) == 0 {
		log.Warn("[%s] empty controller znode, why?", c.participantID)
	} else {
		c.leaderID = string(b)
	}
}

func (c *controller) onBecomingLeader() {
	// rebalance is called when either:
	// 1. participants change
	// 2. resources change
	c.zc.SubscribeChildChanges(c.kb.participants(), c.pcl)
	c.zc.SubscribeChildChanges(c.kb.resources(), c.rcl)

	log.Trace("become controller leader and trigger rebalance!")
	c.rebalance()
}

func (c *controller) rebalance() {
	if !c.zc.IsConnected() {
		log.Trace("[%s] controller disconnected", c.participantID)
		return
	}

	participants, err := c.zc.Children(c.kb.participants())
	if err != nil {
		// TODO
		log.Error("[%s] %s", c.participantID, err)
		return
	}
	if len(participants) == 0 {
		log.Warn("[%s] no alive participants found", c.participantID)
		return
	}

	encodedResources, err := c.zc.Children(c.kb.resources())
	if err != nil {
		log.Error("[%s] %s", c.participantID, err)
		return
	}
	if len(encodedResources) == 0 {
		log.Warn("[%s] no resources found", c.participantID)
		return
	}

	resources := make([]string, len(encodedResources))
	for i, encodedResource := range encodedResources {
		resources[i], err = c.kb.decodeResource(encodedResource)
		if err != nil {
			// should never happen
			log.Critical("%s: %v", encodedResource, err)
			continue
		}
	}

	c.onRebalance(assignResourcesToParticipants(participants, resources))
}
