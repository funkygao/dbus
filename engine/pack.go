package engine

import (
	"fmt"
	"sync/atomic"
)

// PipelinePack is the pipeline data structure that is transferred between plugins.
type PipelinePack struct {
	recycleChan chan *PipelinePack
	refCount    int32

	input bool

	// To avoid infinite message loops
	msgLoopCount int

	// Used internally to stamp diagnostic information
	diagnostics *PacketTracking

	// For routing
	Ident string

	// Project name
	Project string

	Payload []byte
}

func NewPipelinePack(recycleChan chan *PipelinePack) (this *PipelinePack) {
	return &PipelinePack{
		recycleChan:  recycleChan,
		refCount:     int32(1),
		msgLoopCount: 0,
		input:        false,
		diagnostics:  NewPacketTracking(),
	}
}

func (this *PipelinePack) incRef() {
	atomic.AddInt32(&this.refCount, 1)
}

func (this PipelinePack) String() string {
	return fmt.Sprintf("{%s:%s, %+v, %s}", this.Project, this.Ident, this.input, string(this.Payload))
}

func (this *PipelinePack) Reset() {
	this.refCount = int32(1)
	this.msgLoopCount = 0
	this.input = false
	this.diagnostics.Reset()

	this.Project = ""
	this.Ident = ""
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
