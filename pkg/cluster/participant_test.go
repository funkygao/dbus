package cluster

import (
	"testing"

	"github.com/funkygao/assert"
)

func TestParticipant(t *testing.T) {
	p := &Participant{Weight: 5}
	assert.Equal(t, `{"weight":5}`, string(p.Marshal()))
	p.From([]byte(`{"weight":12}`))
	assert.Equal(t, 12, p.Weight)

	p.Endpoint = ""
	assert.Equal(t, false, p.Valid())
	p.Endpoint = "http://a.com"
	assert.Equal(t, false, p.Valid())
	p.Endpoint = "12.1.1.1"
	assert.Equal(t, false, p.Valid())
	p.Endpoint = "12.1.1.1:1222"
	assert.Equal(t, true, p.Valid())
}
