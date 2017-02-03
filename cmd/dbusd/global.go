package main

import (
	"github.com/funkygao/dbus/engine"
)

var (
	globals *engine.GlobalConfigStruct

	BuildID = "unknown" // git version id, passed in from shell
	Version = "unknown"
)
