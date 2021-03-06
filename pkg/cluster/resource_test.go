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
		{Name: "2"},
		{Name: "1"},
		{Name: "3"},
	}
	sorted := Resources(rs)
	sort.Sort(sorted)
	t.Logf("%+v", sorted)
	assert.Equal(t, "1", sorted[0].Name)
	assert.Equal(t, "2", sorted[1].Name)
	assert.Equal(t, "3", sorted[2].Name)
}

func TestResourcesMarshal(t *testing.T) {
	rs := []Resource{
		{Name: "2", InputPlugin: "input2"},
		{Name: "1", InputPlugin: "input1"},
		{Name: "3", InputPlugin: "input3"},
	}
	resources := Resources(rs)
	assert.Equal(t, `[{"input_plugin":"input2","name":"2"},{"input_plugin":"input1","name":"1"},{"input_plugin":"input3","name":"3"}]`,
		string(resources.Marshal()))
}

func TestUnmarshalRPCResources(t *testing.T) {
	rs := UnmarshalRPCResources([]byte(`[{"input_plugin":"in.binlog","name":"local://root:@localhost:3306"}]`))
	assert.Equal(t, "local://root:@localhost:3306", rs[0].Name)
	assert.Equal(t, "in.binlog", rs[0].InputPlugin)
}
