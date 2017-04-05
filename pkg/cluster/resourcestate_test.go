package cluster

import (
	"testing"

	"github.com/funkygao/assert"
)

func TestResourceStateMarshalAndFrom(t *testing.T) {
	rs := NewResourceState()
	assert.Equal(t, 1, rs.Version)

	rs.LeaderEpoch = 5
	rs.Owner = "10.11.1.1:11111"
	assert.Equal(t, `{"leader_epoch":5,"version":1,"owner":"10.11.1.1:11111"}`, string(rs.Marshal()))

	rs.LeaderEpoch = 0
	rs.Owner = ""
	rs.From([]byte(`{"leader_epoch":5,"version":1,"owner":"10.11.1.1:11111"}`))
	assert.Equal(t, 5, rs.LeaderEpoch)
	assert.Equal(t, "10.11.1.1:11111", rs.Owner)
}
