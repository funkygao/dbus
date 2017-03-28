package zk

import (
	"encoding/base64"
	"path"
)

const rootPath = "/dbus/cluster"

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

func (kb *keyBuilder) resources() string {
	return path.Join(rootPath, "resources")
}

func (kb *keyBuilder) resource(resource string) string {
	return path.Join(kb.resources(), kb.encodeResource(resource))
}

func (kb *keyBuilder) encodeResource(resource string) string {
	return base64.URLEncoding.EncodeToString([]byte(resource))
}

func (kb *keyBuilder) decodeResource(encodedResource string) (string, error) {
	b, err := base64.URLEncoding.DecodeString(encodedResource)
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
