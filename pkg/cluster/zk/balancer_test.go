package zk

import (
	"testing"

	"github.com/funkygao/assert"
)

func TestAssignResourcesToParticipants_Normal(t *testing.T) {
	resources := []string{"a", "b", "c", "d", "e"}
	participants := []string{"1", "2"}

	decision := assignResourcesToParticipants(participants, resources)
	t.Logf("%+v", decision)
	assert.Equal(t, 3, len(decision["1"]))
	assert.Equal(t, 2, len(decision["2"]))
}

func TestAssignResourcesToParticipants_EmptyJobQueues(t *testing.T) {
	resources := []string{}
	participants := []string{"1", "2"}

	decision := assignResourcesToParticipants(participants, resources)
	t.Logf("%+v", decision)
}

func TestAssignResourcesToParticipants_ActorMoreThanJobQueues(t *testing.T) {
	resources := []string{"job1"}
	participants := []string{"1", "2"}

	decision := assignResourcesToParticipants(participants, resources)
	t.Logf("%+v", decision)
	assert.Equal(t, 0, len(decision["2"]))
	assert.Equal(t, 1, len(decision["1"]))
}
