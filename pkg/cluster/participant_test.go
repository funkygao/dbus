package cluster

import (
	"sort"
	"testing"

	"github.com/funkygao/assert"
)

func TestParticipant(t *testing.T) {
	p := &Participant{Weight: 5}
	assert.Equal(t, `{"weight":5}`, string(p.Marshal()))
	p.From([]byte(`{"weight":12}`))
	assert.Equal(t, 12, p.Weight)
	p.Revision = "109faeb"
	assert.Equal(t, `{"weight":12,"revision":"109faeb"}`, string(p.Marshal()))

	p.Endpoint = ""
	p.Revision = ""
	assert.Equal(t, false, p.Valid())
	p.Endpoint = "http://a.com"
	assert.Equal(t, false, p.Valid())
	p.Endpoint = "12.1.1.1"
	assert.Equal(t, false, p.Valid())
	p.Endpoint = "12.1.1.1:1222"
	assert.Equal(t, true, p.Valid())

	p = &Participant{}
	p.From([]byte(`{"weight":12,"revision":"109faeb","endpoint":"12.12.12.12:1888"}`))
	assert.Equal(t, 12, p.Weight)
	assert.Equal(t, "109faeb", p.Revision)
	assert.Equal(t, "12.12.12.12:1888", p.Endpoint)
	assert.Equal(t, "http://12.12.12.12:1888", p.RPCEndpoint())
	p.APIPort = 1999
	assert.Equal(t, "http://12.12.12.12:1999", p.APIEndpoint())
}

func TestParticipantsSort(t *testing.T) {
	ps := []Participant{
		{Endpoint: "2"},
		{Endpoint: "1"},
		{Endpoint: "3"},
	}
	sorted := Participants(ps)
	sort.Sort(sorted)
	t.Logf("%+v", sorted)
	assert.Equal(t, "1", sorted[0].Endpoint)
	assert.Equal(t, "2", sorted[1].Endpoint)
	assert.Equal(t, "3", sorted[2].Endpoint)
}
