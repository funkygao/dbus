package engine

import (
	"sync"
	"time"
)

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
