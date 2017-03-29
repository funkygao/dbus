package cluster

// Decision is the cluster leader new assignment of resources to participants.
type Decision map[Participant][]Resource

func MakeDecision() Decision {
	return make(map[Participant][]Resource)
}

func (d Decision) Equals(that Decision) bool {
	if len(d) != len(that) {
		return false
	}

	for p, rs := range that {
		for i, r := range rs {
			if d[p][i] != r {
				return false
			}
		}
	}

	return true
}
