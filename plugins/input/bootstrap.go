// Package input provides input plugins.
package input

import (
	_ "github.com/funkygao/dbus/plugins/input/kafka"
	_ "github.com/funkygao/dbus/plugins/input/mysql"
	_ "github.com/funkygao/dbus/plugins/input/redis"

	// external plugins
	_ "github.com/dbus-plugin/mock-input"
	_ "github.com/dbus-plugin/stream-input"
)
