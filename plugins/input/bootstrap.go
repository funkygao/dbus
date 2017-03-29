// Package input provides input plugins.
package input

import (
	_ "github.com/funkygao/dbus/plugins/input/kafka"
	_ "github.com/funkygao/dbus/plugins/input/mysql"
	_ "github.com/funkygao/dbus/plugins/input/redis"
	_ "github.com/funkygao/dbus/plugins/input/stream"
)
