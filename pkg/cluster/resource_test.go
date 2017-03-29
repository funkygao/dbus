package cluster

import (
	"sort"
	"testing"

	"github.com/funkygao/assert"
)

func TestResource(t *testing.T) {
	r := Resource{InputPlugin: "in.binlog"}
	assert.Equal(t, `{"input_plugin":"in.binlog"}`, string(r.Marshal()))
	r.From([]byte(`{"input_plugin":"changed"}`))
	assert.Equal(t, "changed", r.InputPlugin)
}

func TestResourcesSort(t *testing.T) {
	rs := []Resource{
		Resource{Name: "2"},
		Resource{Name: "1"},
		Resource{Name: "3"},
	}
	sorted := Resources(rs)
	sort.Sort(sorted)
	t.Logf("%+v", sorted)
	assert.Equal(t, "1", sorted[0].Name)
	assert.Equal(t, "2", sorted[1].Name)
	assert.Equal(t, "3", sorted[2].Name)
}
