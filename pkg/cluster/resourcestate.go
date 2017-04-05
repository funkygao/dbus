package cluster

import (
	"encoding/json"
)

type ResourceState struct {
	LeaderEpoch int    `json:"leader_epoch"`
	Version     int    `json:"version"`
	Owner       string `json:"owner"` // participant endpoint
}

func NewResourceState() *ResourceState {
	return &ResourceState{Version: 1}
}

func (rs *ResourceState) From(data []byte) {
	json.Unmarshal(data, rs)
}

func (rs *ResourceState) Marshal() []byte {
	b, _ := json.Marshal(rs)
	return b
}

func (rs *ResourceState) BecomeOrphan() {
	rs.LeaderEpoch = -1
	rs.Owner = ""
}

func (rs *ResourceState) IsOrphan() bool {
	return rs.LeaderEpoch == -1 && len(rs.Owner) == 0
}
