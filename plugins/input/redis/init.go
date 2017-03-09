package redis

import (
	"github.com/funkygao/dbus/engine"
)

var (
	_ engine.Input = &RedisbinlogInput{}
)

func init() {
	engine.RegisterPlugin("RedisbinlogInput", func() engine.Plugin {
		return new(RedisbinlogInput)
	})
}
