// Package input provides input plugins.
package input

import (
	// bootstrap internal input pulugins
	_ "github.com/funkygao/dbus/plugins/input/http"
	_ "github.com/funkygao/dbus/plugins/input/kafka"
	_ "github.com/funkygao/dbus/plugins/input/mysql"

	// bootstrap external plugins
	_ "github.com/dbus-plugin/stream-input"
)
