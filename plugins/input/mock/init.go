package mock

import (
	"github.com/funkygao/dbus/engine"
)

var (
	_ engine.Input     = &MockInput{}
	_ engine.Pauser    = &MockInput{}
	_ engine.Restarter = &MockInput{}
)

func init() {
	engine.RegisterPlugin("MockInput", func() engine.Plugin {
		return new(MockInput)
	})
}
