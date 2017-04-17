// Package input provides input plugins.
package input

import (
	_ "github.com/funkygao/dbus/plugins/input/kafka"
	_ "github.com/funkygao/dbus/plugins/input/mysql"

	// external plugins
	_ "github.com/dbus-plugin/stream-input"
)
