package mock

import (
	"github.com/funkygao/dbus/engine"
)

var (
	_ engine.Input = &StreamInput{}
)

func init() {
	engine.RegisterPlugin("StreamInput", func() engine.Plugin {
		return new(StreamInput)
	})
}
