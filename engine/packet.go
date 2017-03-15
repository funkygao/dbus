package engine

import (
	"fmt"
	"sync/atomic"

	"github.com/funkygao/dbus/pkg/sys"
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
	_padding0   [sys.CacheLineSize / 8]uint64 // avoid false sharing
	recycleChan chan *Packet

	_padding1 [sys.CacheLineSize / 8]uint64
	refCount  int32

	_padding2 [sys.CacheLineSize / 8]uint64
	// Ident is used for routing.
	Ident string

	_padding3 [sys.CacheLineSize / 8]uint64 // TODO [7]uint64 should be enough
	// Metadata is used to hold arbitrary data you wish to include.
	// Engine completely ignores this field and is only to be used for
	// pass-through data.
	Metadata interface{}

	_padding4 [sys.CacheLineSize / 8]uint64
	input     Acker

	_padding5 [sys.CacheLineSize / 8]uint64
	Payload   Payloader

	//	buf     []byte TODO
}

func newPacket(recycleChan chan *Packet) *Packet {
	return &Packet{
		recycleChan: recycleChan,
		refCount:    int32(1),
		Metadata:    nil,
	}
}

func (p *Packet) incRef() *Packet {
	atomic.AddInt32(&p.refCount, 1)
	return p
}

func (p *Packet) String() string {
	return fmt.Sprintf("{%s, %+v, %d, %+v}", p.Ident, p.input, atomic.LoadInt32(&p.refCount), p.Payload)
}

// copyTo will copy itself to another Packet.
func (p *Packet) copyTo(other *Packet) {
	other.Ident = p.Ident
	other.input = p.input
	other.Payload = p.Payload // FIXME clone deep copy
}

func (p *Packet) Reset() {
	p.refCount = int32(1)
	p.Ident = ""
	p.Payload = nil
	p.input = nil
}

// Recycle decrement packet reference count and place it back
// to its recycle pool if possible.
func (p *Packet) Recycle() {
	if atomic.AddInt32(&p.refCount, -1) == 0 {
		p.Reset()

		// reuse this pack to avoid re-alloc
		// if recycleChan is full, will block
		p.recycleChan <- p
	}
}
