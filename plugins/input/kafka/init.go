package kafka

import (
	"github.com/funkygao/dbus/engine"
)

var (
	_ engine.Input = &KafkaInput{}
)

func init() {
	engine.RegisterPlugin("KafkaInput", func() engine.Plugin {
		return new(KafkaInput)
	})
}
