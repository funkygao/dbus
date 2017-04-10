package zk

import (
	"net/url"
	"path"
)

var rootPath = "/dbus/cluster"

type keyBuilder struct {
}

func newKeyBuilder() *keyBuilder {
	return &keyBuilder{}
}

func (kb *keyBuilder) upgrade() string {
	return path.Join(rootPath, "upgrade")
}

func (kb *keyBuilder) participants() string {
	return path.Join(rootPath, "participants")
}

func (kb *keyBuilder) participant(id string) string {
	return path.Join(kb.participants(), id)
}

func (kb *keyBuilder) leader() string {
	return path.Join(rootPath, "leader")
}

func (kb *keyBuilder) leaderEpoch() string {
	return path.Join(rootPath, "leader_epoch")
}

func (kb *keyBuilder) resources() string {
	return path.Join(rootPath, "resources")
}

func (kb *keyBuilder) resource(resource string) string {
	return path.Join(kb.resources(), kb.encodeResource(resource))
}

func (kb *keyBuilder) resourceState(resource string) string {
	return path.Join(kb.resource(resource), "state")
}

func (kb *keyBuilder) encodeResource(resource string) string {
	return url.QueryEscape(resource)
}

func (kb *keyBuilder) decodeResource(encodedResource string) (string, error) {
	b, err := url.QueryUnescape(encodedResource)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (kb *keyBuilder) persistentKeys() []string {
	return []string{
		kb.participants(),
		kb.resources(),
	}
}
