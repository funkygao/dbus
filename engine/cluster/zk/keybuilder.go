package zk

import (
	"path"
)

const rootPath = "/dbus"

type keyBuilder struct {
}

func newKeyBuilder() *keyBuilder {
	return &keyBuilder{}
}

func (kb *keyBuilder) resources() string {
	return path.Join(rootPath, "resources")
}

func (kb *keyBuilder) resource(resource string) string {
	// TODO urlencode?
	return path.Join(kb.resources(), resource)
}

func (kb *keyBuilder) participants() string {
	return path.Join(rootPath, "participants")
}

func (kb *keyBuilder) participant(id string) string {
	return path.Join(kb.participants(), id)
}

func (kb *keyBuilder) persistentKeys() []string {
	return []string{
		kb.resources(),
		kb.participants(),
	}
}
