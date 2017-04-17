package state

import (
	"testing"

	"github.com/funkygao/assert"
	"github.com/funkygao/dbus/pkg/checkpoint"
)

func TestFromBinlog(t *testing.T) {
	s, err := From(checkpoint.SchemeBinlog, "/dbus/checkpoint/myslave/12.12.1.2%3A3334", "in.mysql", []byte(`{"file":"f1","offset":5}`))
	assert.Equal(t, nil, err)
	assert.Equal(t, "f1-5", s.String())
	assert.Equal(t, checkpoint.SchemeBinlog, s.Scheme())
}

func TestFromKafka(t *testing.T) {
	s, err := From(checkpoint.SchemeKafka, "/dbus/checkpoint/myslave/12.12.1.2%3A3334", "in.mysql", []byte(`{"pid":2,"offset":5}`))
	assert.Equal(t, nil, err)
	assert.Equal(t, "2-5", s.String())
}
