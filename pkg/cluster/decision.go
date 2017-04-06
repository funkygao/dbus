package cluster

import (
	"encoding/json"
)

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

// IsAssigned returns if the decision has resources assigned for a participant.
func (d Decision) IsAssigned(p Participant) bool {
	_, present := d[p]
	return present
}

// Close revokes all resources from a participant.
func (d Decision) Close(p Participant) {
	d[p] = nil
}

// MarshalJSON implements json.Marshaler.
func (d Decision) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})
	for p, rs := range d {
		m[p.String()] = rs
	}

	return json.Marshal(m)
}

// UnmarshalJSON implements json.Unmarshaler.
func (d Decision) UnmarshalJSON(data []byte) error {
	v := make(map[string][]Resource)
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	for p, rs := range v {
		participant := Participant{
			Endpoint: p,
		}
		d[participant] = rs
	}
	return nil
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
