package model

import (
	"github.com/Shopify/sarama"
	"github.com/funkygao/dbus/engine"
)

var (
	_ engine.Payloader = String("")
	_ sarama.Encoder   = String("")
)

type String string

func (s String) Length() int {
	return len(s)
}

func (s String) String() string {
	return string(s)
}

func (s String) Encode() ([]byte, error) {
	return []byte(s), nil
}
