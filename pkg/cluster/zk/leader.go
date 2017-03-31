package zk

import (
	log "github.com/funkygao/log4go"
)

func (c *controller) onResigningAsLeader() {
	c.zc.UnsubscribeChildChanges(c.kb.participants(), c.pcl)
	c.zc.UnsubscribeChildChanges(c.kb.resources(), c.rcl)
	c.lastDecision = nil
	c.leaderID = ""
}

func (c *controller) onBecomingLeader() {
	c.zc.SubscribeChildChanges(c.kb.participants(), c.pcl)
	c.zc.SubscribeChildChanges(c.kb.resources(), c.rcl)

	log.Trace("become controller leader and trigger rebalance!")
	c.doRebalance()
}
