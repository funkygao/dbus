package model

import (
	"github.com/Shopify/sarama"
	"github.com/funkygao/dbus/engine"
)

var (
	_ engine.Payloader = Bytes{}
	_ sarama.Encoder   = &Bytes{}
)

type Bytes []byte

func (b Bytes) Length() int {
	return len(b)
}

func (b Bytes) String() string {
	return string(b)
}

func (b Bytes) Encode() ([]byte, error) {
	return b, nil
}
