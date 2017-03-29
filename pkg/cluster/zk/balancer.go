package zk

import (
	"sort"

	log "github.com/funkygao/log4go"
)

/// rebalance happens when:
// 1. participants change
// 2. resources change
// 3. becoming leader
func (c *controller) doRebalance() {
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

func assignResourcesToParticipants(participants []string, resources []string) (decision map[string][]string) {
	decision = make(map[string][]string)

	rLen, pLen := len(resources), len(participants)
	if pLen == 0 || rLen == 0 {
		return
	}

	sort.Strings(participants)
	sort.Strings(resources)

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
				decision[participants[pid]] = make([]string, 0)
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
