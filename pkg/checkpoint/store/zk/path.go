package zk

import (
	"path"
)

func realPath(zpath string) string {
	return path.Join("/dbus/checkpoint", zpath)
}
