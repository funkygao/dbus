package cluster

import (
	"testing"

	"github.com/funkygao/assert"
)

func TestDecisionEquals(t *testing.T) {
	d1 := MakeDecision()
	d2 := MakeDecision()
	assert.Equal(t, true, d1.Equals(d2))

	t.SkipNow()
}
