package engine

import (
	"fmt"
	"runtime"
	"time"

	"github.com/funkygao/golib/gofmt"
)

type EngineStats struct {
	engine   *Engine
	MemStats *runtime.MemStats
}

func newEngineStats(e *Engine) (this *EngineStats) {
	this = new(EngineStats)
	this.engine = e
	this.MemStats = new(runtime.MemStats)
	return
}

func (this *EngineStats) Runtime() map[string]interface{} {
	this.refreshMemStats()

	s := make(map[string]interface{})
	s["goroutines"] = runtime.NumGoroutine()
	s["memory.allocated"] = gofmt.ByteSize(this.MemStats.Alloc).String()
	s["memory.mallocs"] = gofmt.ByteSize(this.MemStats.Mallocs).String()
	s["memory.frees"] = gofmt.ByteSize(this.MemStats.Frees).String()
	s["memory.last_gc"] = this.MemStats.LastGC
	s["memory.gc.num"] = this.MemStats.NumGC
	s["memory.gc.num_per_second"] = float64(this.MemStats.NumGC) / time.Since(Globals().StartedAt).Seconds()
	s["memory.gc.total_pause"] = fmt.Sprintf("%dms",
		this.MemStats.PauseTotalNs/uint64(time.Millisecond))
	s["memory.heap.alloc"] = gofmt.ByteSize(this.MemStats.HeapAlloc).String()       // 堆上目前分配的内存
	s["memory.heap.sys"] = gofmt.ByteSize(this.MemStats.HeapSys).String()           // 程序向操作系统申请的内存
	s["memory.heap.idle"] = gofmt.ByteSize(this.MemStats.HeapIdle).String()         // 堆上目前没有使用的内存
	s["memory.heap.released"] = gofmt.ByteSize(this.MemStats.HeapReleased).String() // 回收到操作系统的内存
	s["memory.heap.objects"] = gofmt.Comma(int64(this.MemStats.HeapObjects))
	s["memory.stack"] = gofmt.ByteSize(this.MemStats.StackInuse).String()
	gcPausesMs := make([]string, 0, 20)
	for _, pauseNs := range this.MemStats.PauseNs {
		if pauseNs == 0 {
			continue
		}

		pauseStr := fmt.Sprintf("%dms",
			pauseNs/uint64(time.Millisecond))
		if pauseStr == "0ms" {
			continue
		}

		gcPausesMs = append(gcPausesMs, pauseStr)
	}
	s["memory.gc.pauses"] = gcPausesMs

	return s
}

func (this *EngineStats) refreshMemStats() {
	runtime.ReadMemStats(this.MemStats)
}
