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
	Encode() ([]byte, error)

	// String return the string format of the payload.
	// Useful for debugging.
	String() string
}

// PipelinePack is the pipeline data structure that is transferred between plugins.
type PipelinePack struct {
	recycleChan chan *PipelinePack
	refCount    int32

	input bool

	// For routing
	Ident string

	Payload Payloader
}

func NewPipelinePack(recycleChan chan *PipelinePack) *PipelinePack {
	return &PipelinePack{
		recycleChan: recycleChan,
		refCount:    int32(1),
		input:       false,
	}
}

func (p *PipelinePack) incRef() {
	atomic.AddInt32(&p.refCount, 1)
}

func (p PipelinePack) String() string {
	return fmt.Sprintf("{%s, %+v, %s}", p.Ident, p.input, p.Payload)
}

func (p *PipelinePack) Reset() {
	p.refCount = int32(1)
	p.input = false
	p.Ident = ""
	p.Payload = nil
}

func (p *PipelinePack) Recycle() {
	count := atomic.AddInt32(&p.refCount, -1)
	if count == 0 {
		p.Reset()

		// reuse this pack to avoid re-alloc
		p.recycleChan <- p
	} else if count < 0 {
		fmt.Println("reference count below zero")
	}
}
