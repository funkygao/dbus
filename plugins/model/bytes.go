package model

import (
	"github.com/funkygao/dbus/engine"
)

var _ engine.Payloader = Bytes{}

type Bytes []byte

func (b Bytes) Length() int {
	return len(b)
}

func (b Bytes) String() string {
	return string(b)
}

func (b Bytes) Bytes() []byte {
	return b
}
