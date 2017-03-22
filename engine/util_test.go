package engine

import (
	"testing"

	"github.com/funkygao/assert"
)

func TestParseConfigPath(t *testing.T) {
	zkSvr, path := parseConfigPath("../xx.cf")
	assert.Equal(t, "", zkSvr)
	assert.Equal(t, "../xx.cf", path)

	zkSvr, path = parseConfigPath("localhost:2181/foo/bar")
	assert.Equal(t, "localhost:2181", zkSvr)
	assert.Equal(t, "/foo/bar", path)
}
