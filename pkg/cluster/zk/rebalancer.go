package zk

import (
	log "github.com/funkygao/log4go"
)

/// rebalance happens when:
// 1. participants change
// 2. resources change
// 3. becoming leader
func (c *controller) doRebalance() {
	participants, err := c.LiveParticipants()
	if err != nil {
		// TODO
		log.Critical("[%s] %s", c.participant, err)
		return
	}
	if len(participants) == 0 {
		log.Critical("[%s] no alive participants found", c.participant)
		return
	}

	resources, err := c.RegisteredResources()
	if err != nil {
		// TODO
		log.Critical("[%s] %s", c.participant, err)
		return
	}

	newDecision := c.strategyFunc(participants, resources)
	if !newDecision.Equals(c.lastDecision) {
		c.lastDecision = newDecision

		c.onRebalance(newDecision)
	}
}
