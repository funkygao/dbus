package cluster

import (
	"testing"

	"github.com/funkygao/assert"
)

func TestResource(t *testing.T) {
	r := Resource{InputPlugin: "in.binlog"}
	assert.Equal(t, `{"input_plugin":"in.binlog"}`, string(r.Marshal()))
	r.From([]byte(`{"input_plugin":"changed"}`))
	assert.Equal(t, "changed", r.InputPlugin)
}
