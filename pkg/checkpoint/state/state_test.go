package state

import (
	"testing"

	"github.com/funkygao/assert"
	"github.com/funkygao/dbus/pkg/checkpoint"
)

func TestLoadBinlog(t *testing.T) {
	s, err := Load(checkpoint.SchemeBinlog, "/dbus/checkpoint/myslave/12.12.1.2%3A3334", []byte(`{"file":"f1","offset":5}`))
	assert.Equal(t, nil, err)
	assert.Equal(t, "f1-5", s.String())
	assert.Equal(t, checkpoint.SchemeBinlog, s.Scheme())
}

func TestLoadKafka(t *testing.T) {
	s, err := Load(checkpoint.SchemeKafka, "/dbus/checkpoint/myslave/12.12.1.2%3A3334", []byte(`{"pid":2,"offset":5}`))
	assert.Equal(t, nil, err)
	assert.Equal(t, "2-5", s.String())
}
