package zk

import (
	"path"

	"github.com/funkygao/dbus/pkg/checkpoint"
)

const root = "/dbus/checkpoint"

func realPath(state checkpoint.State, zpath string) string {
	return path.Join(root, state.Scheme(), zpath)
}
