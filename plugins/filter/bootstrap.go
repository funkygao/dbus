// Package filter provides filter plugins.
package filter

import (
	_ "github.com/funkygao/dbus/plugins/filter/mysql"

	// external plugins
	_ "github.com/dbus-plugin/mock-filter"
)
