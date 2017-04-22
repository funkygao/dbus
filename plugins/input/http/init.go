package http

import (
	"github.com/funkygao/dbus/engine"
)

var (
	_ engine.Input = &HTTPInput{}
)

func init() {
	engine.RegisterPlugin("HTTPInput", func() engine.Plugin {
		return new(HTTPInput)
	})
}
