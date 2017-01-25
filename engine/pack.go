package engine

import (
	"sync"
	"sync/atomic"
	"time"
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

func (this *PipelinePack) CopyTo(that *PipelinePack) {
	that.Project = this.Project
	that.Ident = this.Ident
	that.Input = this.Input
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

type PacketTracking struct {
	LastAccess time.Time
	mutex      sync.Mutex

	// Records the plugins the packet has been handed to
	pluginRunners []PluginRunner
}

func NewPacketTracking() *PacketTracking {
	return &PacketTracking{LastAccess: time.Now(),
		mutex:         sync.Mutex{},
		pluginRunners: make([]PluginRunner, 0, 8)}
}

func (this *PacketTracking) AddStamp(pluginRunner PluginRunner) {
	this.mutex.Lock()
	this.pluginRunners = append(this.pluginRunners, pluginRunner)
	this.mutex.Unlock()
	this.LastAccess = time.Now()
}

func (this *PacketTracking) Reset() {
	this.mutex.Lock()
	this.pluginRunners = this.pluginRunners[:0] // a tip in golang to avoid re-alloc
	this.mutex.Unlock()
	this.LastAccess = time.Now()
}

func (this *PacketTracking) Runners() []PluginRunner {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return this.pluginRunners
}

func (this *PacketTracking) RunnerCount() int {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return len(this.pluginRunners)

}
