package zk

import (
	"testing"

	//"github.com/funkygao/assert"
	"github.com/funkygao/dbus/pkg/cluster"
)

func TestAssignResourcesToParticipants_Normal(t *testing.T) {
	resources := []cluster.Resource{
		cluster.Resource{Name: "a"},
		cluster.Resource{Name: "b"},
		cluster.Resource{Name: "c"},
		cluster.Resource{Name: "d"},
		cluster.Resource{Name: "e"},
	}
	participants := []cluster.Participant{
		cluster.Participant{Endpoint: "1"},
		cluster.Participant{Endpoint: "2"},
	}

	decision := assignResourcesToParticipants(participants, resources)
	t.Logf("%+v", decision)
}
