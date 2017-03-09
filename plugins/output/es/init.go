package es

import (
	"github.com/funkygao/dbus/engine"
)

var (
	_ engine.Output = &ESOutput{}
)

func init() {
	engine.RegisterPlugin("ESOutput", func() engine.Plugin {
		return new(ESOutput)
	})
}
