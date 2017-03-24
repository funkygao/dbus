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

func (kb *keyBuilder) participants() string {
	return path.Join(rootPath, "participants")
}

func (kb *keyBuilder) participant(id string) string {
	return path.Join(kb.participants(), id)
}

func (kb *keyBuilder) controller() string {
	return path.Join(rootPath, "controller")
}

func (kb *keyBuilder) persistentKeys() []string {
	return []string{
		kb.participants(),
	}
}
