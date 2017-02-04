package main

import (
	"github.com/funkygao/dbus/engine"
)

var (
	globals *engine.GlobalConfig

	BuildID = "unknown" // git version id, passed in from shell
	Version = "unknown"
)
