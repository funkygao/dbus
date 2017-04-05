package cluster

import (
	"encoding/json"
)

// Resource is a assignable entity in a cluster.
type Resource struct {
	InputPlugin string `json:"input_plugin,omitempty"`
	Name        string `json:"name,omitempty"`
	Cost        int    `json:"cost,omitempty"`

	// will not persist in json
	State *ResourceState `json:"-"`
}

// UnmarshalRPCResources is used by a participant to unmarshal RPC reblance request body into list of resources.
// It will get all the resources that the participant will lead.
func UnmarshalRPCResources(data []byte) []Resource {
	r := make([]Resource, 0)
	json.Unmarshal(data, &r)
	return r
}

func (r *Resource) DSN() string {
	return r.Name
}

func (r *Resource) Marshal() []byte {
	b, _ := json.Marshal(r)
	return b
}

func (r Resource) Equals(that Resource) bool {
	return r.InputPlugin == that.InputPlugin && r.Name == that.Name
}

func (r *Resource) From(data []byte) {
	json.Unmarshal(data, r)
}

type Resources []Resource

func (rs Resources) Marshal() []byte {
	b, _ := json.Marshal(rs)
	return b
}

func (rs Resources) Len() int {
	return len(rs)
}

func (rs Resources) Less(i, j int) bool {
	return rs[i].Name < rs[j].Name
}

func (rs Resources) Swap(i, j int) {
	rs[i], rs[j] = rs[j], rs[i]
}
