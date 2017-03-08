package engine

import (
	"fmt"
	"sync/atomic"
)

// Payloader defines the contract of Packet payload.
// Any plugin transferrable data must implement this interface.
type Payloader interface {
	// Length returns the size of the payload in bytes.
	Length() int

	// Bytes returns the marshalled byte array of the payload.
	Encode() ([]byte, error)
}

// Packet is the pipeline data structure that is transferred between plugins.
type Packet struct {
	_padding0   [8]uint64 // avoid false sharing
	recycleChan chan *Packet

	_padding1 [8]uint64
	refCount  int32

	_padding2 [8]uint64
	input     bool // TODO kill this

	_padding3 [8]uint64
	// Ident is used for routing.
	Ident string

	_padding4 [8]uint64 // TODO [7]uint64 should be enough
	// Metadata is used to hold arbitrary data you wish to include.
	// Engine completely ignores this field and is only to be used for
	// pass-through data.
	Metadata interface{}

	_padding5 [8]uint64
	Payload   Payloader
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

// CopyTo will copy itself to another Packet.
func (p *Packet) CopyTo(other *Packet) {
	other.input = p.input
	other.Ident = p.Ident
	other.Payload = p.Payload // FIXME clone deep copy
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
		// if recycleChan is full, will block
		p.recycleChan <- p
	}
}
