package zk

import (
	"net/url"
	"path"

	"github.com/funkygao/dbus/pkg/checkpoint"
)

var root = "/dbus/checkpoint"

func realPath(state checkpoint.State, zpath string) string {
	return path.Join(root, state.Scheme(), url.QueryEscape(zpath))
}
