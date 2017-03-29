package zk

import (
	//"sort"

	"github.com/funkygao/dbus/pkg/cluster"
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

	c.onRebalance(assignResourcesToParticipants(participants, resources))
}

func assignResourcesToParticipants(participants []cluster.Participant, resources []cluster.Resource) (decision cluster.Decision) {
	decision = cluster.MakeDecision()

	rLen, pLen := len(resources), len(participants)
	if pLen == 0 || rLen == 0 {
		// FIXME if pLen>0 && rLen==0, should not return
		return
	}

	/*
		sort.Strings(participants)
		sort.Strings(resources)*/

	nResourcesPerParticipant, nparticipantsWithExtraResource := rLen/pLen, rLen%pLen

	for pid := 0; pid < pLen; pid++ {
		extraN := 1
		if pid+1 > nparticipantsWithExtraResource {
			extraN = 0
		}

		nResources := nResourcesPerParticipant + extraN
		startResourceIdx := nResourcesPerParticipant*pid + min(pid, nparticipantsWithExtraResource)
		for j := startResourceIdx; j < startResourceIdx+nResources; j++ {
			if _, present := decision[participants[pid]]; !present {
				decision[participants[pid]] = make([]cluster.Resource, 0)
			}
			decision[participants[pid]] = append(decision[participants[pid]], resources[j])
		}
	}
	return
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
