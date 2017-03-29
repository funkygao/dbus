package cluster

import (
	"testing"

	"github.com/funkygao/assert"
)

func TestStategyRoundRobin(t *testing.T) {
	resources := []Resource{
		{Name: "a"},
		{Name: "b"},
		{Name: "c"},
		{Name: "d"},
		{Name: "e"},
	}
	p1 := Participant{Endpoint: "1"}
	p2 := Participant{Endpoint: "2"}
	participants := []Participant{
		p1,
		p2,
	}

	decision := GetStrategyFunc(StrategyRoundRobin)(participants, resources)

	// p1: a,b,c
	// p2: d,e
	t.Logf("%+v", decision[p1])
	t.Logf("%+v", decision[p2])

	assert.Equal(t, 3, len(decision[p1]))
	assert.Equal(t, 2, len(decision[p2]))
}
