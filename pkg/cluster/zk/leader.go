package zk

import (
	log "github.com/funkygao/log4go"
)

func (c *controller) onResigningAsLeader() {

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
