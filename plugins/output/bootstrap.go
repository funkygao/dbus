// Package output provides output plugins.
package output

import (
	_ "github.com/funkygao/dbus/plugins/output/es"
	_ "github.com/funkygao/dbus/plugins/output/kafka"

	// external plugins
	_ "github.com/dbus-plugin/mock-output"
)
