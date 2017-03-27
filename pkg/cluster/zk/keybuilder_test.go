package zk

import (
	"strings"
	"testing"

	"github.com/funkygao/assert"
)

func TestKeyBuilder(t *testing.T) {
	kb := newKeyBuilder()
	assert.Equal(t, "/dbus/participants", kb.participants())
	assert.Equal(t, "/dbus/participants/12.11.11.11-9876", kb.participant("12.11.11.11-9876"))
	assert.Equal(t, true, strings.HasPrefix(kb.participant("foobar"), kb.participants()))
}
