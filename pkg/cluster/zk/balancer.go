package zk

import (
	"sort"
)

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
