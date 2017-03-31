package cluster

import (
	"sort"

	"github.com/funkygao/golib/math"
)

func assignRoundRobin(participants []Participant, resources []Resource) (decision Decision) {
	sortedParticipants := Participants(participants)
	sortedResources := Resources(resources)
	sort.Sort(sortedParticipants)
	sort.Sort(sortedResources)

	rLen, pLen := len(resources), len(participants)
	nResourcesPerParticipant, nparticipantsWithExtraResource := rLen/pLen, rLen%pLen

	decision = MakeDecision()
	for pid := 0; pid < pLen; pid++ {
		extraN := 1
		if pid+1 > nparticipantsWithExtraResource {
			extraN = 0
		}

		nResources := nResourcesPerParticipant + extraN
		startResourceIdx := nResourcesPerParticipant*pid + math.MinInt(pid, nparticipantsWithExtraResource)
		for j := startResourceIdx; j < startResourceIdx+nResources; j++ {
			decision.Assign(sortedParticipants[pid], sortedResources[j])
		}
	}

	// notify the idle participants
	for _, p := range participants {
		if !decision.IsAssigned(p) {
			decision.Close(p)
		}
	}

	return
}
