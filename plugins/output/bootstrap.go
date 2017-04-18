// Package output provides output plugins.
package output

import (
	// bootstrap internal output plugins
	_ "github.com/funkygao/dbus/plugins/output/es"
	_ "github.com/funkygao/dbus/plugins/output/kafka"
)
