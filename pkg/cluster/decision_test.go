package cluster

import (
	"testing"

	"github.com/funkygao/assert"
)

func TestDecisionAssign(t *testing.T) {
	d := MakeDecision()
	p1 := Participant{Endpoint: "p1"}
	p2 := Participant{Endpoint: "p2"}
	r1 := Resource{Name: "r1"}
	r2 := Resource{Name: "r2"}
	r3 := Resource{Name: "r3"}

	d.Assign(p1, r1)
	assert.Equal(t, 1, len(d.Get(p1)))
	d.Assign(p1, r2)
	assert.Equal(t, 2, len(d.Get(p1)))
	d.Assign(p1, r3)
	assert.Equal(t, 3, len(d.Get(p1)))

	assert.Equal(t, 0, len(d.Get(p2)))
	d.Assign(p2, r1)
	assert.Equal(t, 1, len(d.Get(p2)))
	d.Assign(p2, r2)
	assert.Equal(t, 2, len(d.Get(p2)))
}

func TestDecisionEquals(t *testing.T) {
	d1 := MakeDecision()
	d2 := MakeDecision()
	assert.Equal(t, true, d1.Equals(d2))

	p1 := Participant{Endpoint: "p1"}
	p2 := Participant{Endpoint: "p2"}
	r1 := Resource{Name: "r1"}
	r2 := Resource{Name: "r2"}
	r3 := Resource{Name: "r3"}

	d1[p1] = []Resource{r1}
	assert.Equal(t, false, d1.Equals(d2))

	d2[p1] = []Resource{r1}
	assert.Equal(t, true, d1.Equals(d2))

	d1[p1] = []Resource{r2}
	assert.Equal(t, false, d1.Equals(d2))

	d1 = MakeDecision()
	d2 = MakeDecision()
	d1.Assign(p1, r1, r2)
	d2.Assign(p2, r1, r2, r3)
	assert.Equal(t, 3, len(d2[p2]))
	assert.Equal(t, 0, len(d2[p1]))
	assert.Equal(t, false, d1.Equals(d2))

	d1 = MakeDecision()
	d2 = MakeDecision()
	d1.Assign(p1, r1, r2)
	d1.Assign(p2, r3)
	d2.Assign(p2, r1)
	d2.Assign(p2, r2, r3)
	t.Logf("%+v", d1)
	assert.Equal(t, false, d1.Equals(d2))
}
