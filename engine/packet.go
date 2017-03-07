package engine

import (
	"fmt"
	"sync/atomic"
)

//go:generate structlayout github.com/funkygao/dbus/engine Packet

// Payloader defines the contract of Packet payload.
// Any plugin transferrable data must implement this interface.
type Payloader interface {
	// Length returns the size of the payload in bytes.
	Length() int

	// Bytes returns the marshalled byte array of the payload.
	Encode() ([]byte, error)

	// String return the string format of the payload.
	// Useful for debugging.
	String() string
}

// Packet is the pipeline data structure that is transferred between plugins.
// TODO padding, false sharing
type Packet struct {
	recycleChan chan *Packet

	refCount int32
	input    bool

	// For routing
	Ident string

	Payload Payloader
	//	buf     []byte TODO
}

func NewPacket(recycleChan chan *Packet) *Packet {
	return &Packet{
		recycleChan: recycleChan,
		refCount:    int32(1),
		input:       false,
	}
}

func (p *Packet) incRef() {
	atomic.AddInt32(&p.refCount, 1)
}

func (p Packet) String() string {
	return fmt.Sprintf("{%s, %+v, %s}", p.Ident, p.input, p.Payload)
}

func (p *Packet) Reset() {
	p.refCount = int32(1)
	p.input = false
	p.Ident = ""
	p.Payload = nil
}

func (p *Packet) Recycle() {
	if atomic.AddInt32(&p.refCount, -1) == 0 {
		p.Reset()

		// reuse this pack to avoid re-alloc
		p.recycleChan <- p
	}
}
