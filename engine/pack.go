package engine

import (
	"fmt"
	"sync/atomic"
)

// Payloader defines the contract of PipelinePack payload.
// Any plugin transferrable data must implement this interface.
type Payloader interface {
	// Length returns the size of the payload in bytes.
	Length() int

	// Bytes returns the marshalled byte array of the payload.
	Bytes() []byte

	// String return the string format of the payload.
	// Useful for debugging.
	String() string
}

// PipelinePack is the pipeline data structure that is transferred between plugins.
type PipelinePack struct {
	recycleChan chan *PipelinePack
	refCount    int32

	input bool

	// Used internally to stamp diagnostic information
	diagnostics *PacketTracking

	// For routing
	Ident string

	Payload Payloader
}

func NewPipelinePack(recycleChan chan *PipelinePack) (this *PipelinePack) {
	return &PipelinePack{
		recycleChan: recycleChan,
		refCount:    int32(1),
		input:       false,
		diagnostics: NewPacketTracking(),
	}
}

func (this *PipelinePack) incRef() {
	atomic.AddInt32(&this.refCount, 1)
}

func (this PipelinePack) String() string {
	return fmt.Sprintf("{%s, %+v, %s}", this.Ident, this.input, this.Payload)
}

func (this *PipelinePack) Reset() {
	this.refCount = int32(1)
	this.input = false
	this.diagnostics.Reset()

	this.Ident = ""
	this.Payload = nil
}

func (this *PipelinePack) Recycle() {
	count := atomic.AddInt32(&this.refCount, -1)
	if count == 0 {
		this.Reset()

		// reuse this pack to avoid re-alloc
		this.recycleChan <- this
	} else if count < 0 {
		Globals().Panic("reference count below zero")
	}
}
