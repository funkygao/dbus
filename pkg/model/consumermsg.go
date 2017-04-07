package model

import (
	"github.com/Shopify/sarama"
	"github.com/funkygao/dbus/engine"
)

var (
	_ engine.Payloader = &ConsumerMessage{}
)

// ConsumerMessage is a kafka consumer messsage that is also a payloader.
type ConsumerMessage struct {
	*sarama.ConsumerMessage
}

func (m ConsumerMessage) Length() int {
	return len(m.Value)
}

func (m ConsumerMessage) Encode() ([]byte, error) {
	return m.Value, nil
}
