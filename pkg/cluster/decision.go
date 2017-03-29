package cluster

// Decision is the cluster leader new assignment of resources to participants.
type Decision map[Participant][]Resource

func MakeDecision() Decision {
	return make(map[Participant][]Resource)
}
