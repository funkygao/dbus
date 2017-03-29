package cluster

import (
	"encoding/json"
)

// Resource is a assignable entity in a cluster.
type Resource struct {
	InputPlugin string `json:"input_plugin,omitempty"`
	Name        string `json:"name,omitempty"`
}

func (r *Resource) Marshal() []byte {
	b, _ := json.Marshal(r)
	return b
}

func (r *Resource) From(data []byte) {
	json.Unmarshal(data, r)
}
