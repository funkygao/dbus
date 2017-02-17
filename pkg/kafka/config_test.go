package kafka

import (
	"testing"

	"github.com/funkygao/assert"
)

func TestGenerateClientID(t *testing.T) {
	id, err := generateClientID()
	assert.Equal(t, nil, err)
	t.Logf("%s", id)
}

func TestDefaultConfig(t *testing.T) {
	c := DefaultConfig()
	assert.Equal(t, true, c.async)
	assert.Equal(t, false, c.SyncMode().async)
	assert.Equal(t, true, c.AsyncMode().async)
}
