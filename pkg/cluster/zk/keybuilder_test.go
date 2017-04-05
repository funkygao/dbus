package zk

import (
	"strings"
	"testing"

	"github.com/funkygao/assert"
)

func TestKeyBuilder(t *testing.T) {
	kb := newKeyBuilder()

	assert.Equal(t, "/dbus/cluster/upgrade", kb.upgrade())

	// participant related
	assert.Equal(t, "/dbus/cluster/participants", kb.participants())
	assert.Equal(t, "/dbus/cluster/participants/12.11.11.11-9876", kb.participant("12.11.11.11-9876"))
	assert.Equal(t, true, strings.HasPrefix(kb.participant("foobar"), kb.participants()))

	// resource related
	assert.Equal(t, "/dbus/cluster/resources/bG9jYWw6Ly9yb290OkBsb2NhbGhvc3Q6MzMwNg==", kb.resource("local://root:@localhost:3306"))
	assert.Equal(t, "/dbus/cluster/resources/bG9jYWw6Ly9yb290OkBsb2NhbGhvc3Q6MzMwNg==/state", kb.resourceState("local://root:@localhost:3306"))

	// controller related
	assert.Equal(t, "/dbus/cluster/leader", kb.leader())
	assert.Equal(t, "/dbus/cluster/leader_epoch", kb.leaderEpoch())
}

func TestKeyBuildEncodeDecodeResource(t *testing.T) {
	kb := newKeyBuilder()
	encoded := "bG9jYWw6Ly9yb290OkBsb2NhbGhvc3Q6MzMwNg=="
	resource := "local://root:@localhost:3306"
	assert.Equal(t, encoded, kb.encodeResource(resource))
	r, err := kb.decodeResource(encoded)
	assert.Equal(t, nil, err)
	assert.Equal(t, resource, r)
}
