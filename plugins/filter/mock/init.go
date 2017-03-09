package mock

import (
	"github.com/funkygao/dbus/engine"
)

var (
	_ engine.Filter = &MockFilter{}
)

func init() {
	engine.RegisterPlugin("MockFilter", func() engine.Plugin {
		return new(MockFilter)
	})
}
