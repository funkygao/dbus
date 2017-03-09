package kafka

import (
	"github.com/funkygao/dbus/engine"
)

var (
	_ engine.Output = &KafkaOutput{}
)

func init() {
	engine.RegisterPlugin("KafkaOutput", func() engine.Plugin {
		return new(KafkaOutput)
	})
}
