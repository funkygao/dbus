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

// KeyValuer is an interface that can be applied on Payloader.
type KeyValuer interface {

	// Get returns value of the key.
	Get(k string) (v interface{}, ok bool)

	// Set stores a value for the key.
	Set(k string, v interface{})
}

// Packet is the pipeline data structure that is transferred between plugins.
//
// TODO hide it to private.
type Packet struct {
	recycleChan chan *Packet

	refCount int32
	acker    Acker // the Input it originates from
	// buf []byte TODO reuse memory

	// Ident is used for routing.
	Ident string

	// Metadata is used to hold arbitrary data you wish to include.
	// Engine completely ignores this field and is only to be used for
	// pass-through data.
	Metadata interface{}

	Payload Payloader
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
	return fmt.Sprintf("{%s, %+v, %d, %+v}", p.Ident, p.acker, atomic.LoadInt32(&p.refCount), p.Payload)
}

// copyTo will copy itself to another Packet.
func (p *Packet) copyTo(other *Packet) {
	other.Ident = p.Ident
	other.acker = p.acker
	other.Payload = p.Payload // FIXME clone deep copy
}

func (p *Packet) Reset() {
	p.refCount = int32(1)
	p.Ident = ""
	p.Payload = nil
	p.acker = nil
}

// ack notifies the Packet's source Input that it is successfully processed.
// ack is called by Output plugin.
func (p *Packet) ack() error {
	return p.acker.Ack(p)
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
