package cluster

// Decision is the cluster leader new assignment of resources to participants.
type Decision map[Participant][]Resource

// MakeDecision creates an empty decision.
func MakeDecision() Decision {
	return make(map[Participant][]Resource)
}

// Assign assigns some resource to a participant.
func (d Decision) Assign(p Participant, rs ...Resource) {
	if _, present := d[p]; !present {
		d[p] = rs
	} else {
		d[p] = append(d[p], rs...)
	}
}

// Get returns all the assigned resources of a participant.
func (d Decision) Get(p Participant) []Resource {
	return d[p]
}

func (d Decision) IsAssigned(p Participant) bool {
	_, present := d[p]
	return present
}

func (d Decision) Close(p Participant) {
	d[p] = nil
}

// Equals compares 2 decision, return true if they are the same.
func (d Decision) Equals(that Decision) bool {
	if len(d) != len(that) {
		return false
	}

	for p, rs := range that {
		if len(rs) != len(d[p]) {
			return false
		}

		for i, r := range rs {
			if d[p][i] != r {
				return false
			}
		}
	}

	return true
}
