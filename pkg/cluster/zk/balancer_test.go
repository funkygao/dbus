package zk

import (
	"testing"

	"github.com/funkygao/assert"
	"github.com/funkygao/dbus/pkg/cluster"
)

func TestAssignResourcesToParticipants_Normal(t *testing.T) {
	resources := []cluster.Resource{
		{Name: "a"},
		{Name: "b"},
		{Name: "c"},
		{Name: "d"},
		{Name: "e"},
	}
	p1 := cluster.Participant{Endpoint: "1"}
	p2 := cluster.Participant{Endpoint: "2"}
	participants := []cluster.Participant{
		p1,
		p2,
	}

	decision := assignResourcesToParticipants(participants, resources)

	// p1: a,b,c
	// p2: d,e
	t.Logf("%+v", decision[p1])
	t.Logf("%+v", decision[p2])

	assert.Equal(t, 3, len(decision[p1]))
	assert.Equal(t, 2, len(decision[p2]))
}
