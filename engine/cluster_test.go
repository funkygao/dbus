package engine

import (
	"testing"

	"github.com/funkygao/assert"
)

func TestEncodeResources(t *testing.T) {
	e := &Engine{}
	assert.Equal(t, []string{"bin.input:local://root:@localhost:3306"}, e.encodeResources("bin.input", []string{"local://root:@localhost:3306"}))
}
