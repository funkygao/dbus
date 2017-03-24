package zk

import (
	"strings"
	"testing"

	"github.com/funkygao/assert"
)

func TestKeyBuilder(t *testing.T) {
	kb := newKeyBuilder()
	assert.Equal(t, "/dbus/participants", kb.participants())
	assert.Equal(t, "/dbus/resources", kb.resources())
	assert.Equal(t, "/dbus/resources/localhost:3306", kb.resource("localhost:3306"))
	assert.Equal(t, true, strings.HasPrefix(kb.resource("foobar"), kb.resources()))
}
