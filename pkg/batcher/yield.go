package batcher

import (
	"runtime"
	"time"
)

const backoff = time.Microsecond

var yield func()

func yieldWithSleep() {
	time.Sleep(backoff)
}

func yieldWithGoSched() {
	runtime.Gosched()
}

func init() {
	yield = yieldWithSleep
}
