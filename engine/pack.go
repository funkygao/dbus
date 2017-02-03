package engine

import (
	"sync/atomic"
)

// PipelinePack is the pipeline data structure that is transfered between plugins.
type PipelinePack struct {
	RecycleChan chan *PipelinePack
	RefCount    int32

	Input bool

	// For routing
	Ident string

	// Project name
	Project string

	// To avoid infinite message loops
	MsgLoopCount int

	Payload []byte

	// Used internally to stamp diagnostic information
	diagnostics *PacketTracking
}

func NewPipelinePack(recycleChan chan *PipelinePack) (this *PipelinePack) {
	return &PipelinePack{
		RecycleChan:  recycleChan,
		RefCount:     int32(1),
		MsgLoopCount: 0,
		Input:        false,
		diagnostics:  NewPacketTracking(),
	}
}

func (this *PipelinePack) IncRef() {
	atomic.AddInt32(&this.RefCount, 1)
}

func (this *PipelinePack) Reset() {
	this.RefCount = int32(1)
	this.MsgLoopCount = 0
	this.Project = ""
	this.Input = false
	this.Ident = ""
	this.diagnostics.Reset()
}

func (this *PipelinePack) Recycle() {
	count := atomic.AddInt32(&this.RefCount, -1)
	if count == 0 {
		this.Reset()

		// reuse this pack to avoid re-alloc
		this.RecycleChan <- this
	} else if count < 0 {
		Globals().Panic("reference count below zero")
	}
}
