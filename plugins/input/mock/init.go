package mock

import (
	"github.com/funkygao/dbus/engine"
)

var (
	_ engine.Input = &MockInput{}
)

func init() {
	engine.RegisterPlugin("MockInput", func() engine.Plugin {
		return new(MockInput)
	})
}
