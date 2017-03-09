package mock

import (
	"github.com/funkygao/dbus/engine"
)

var (
	_ engine.Output = &MockOutput{}
)

func init() {
	engine.RegisterPlugin("MockOutput", func() engine.Plugin {
		return new(MockOutput)
	})
}
