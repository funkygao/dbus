package batcher

import (
	"runtime"
	"time"
)

const backoff = time.Microsecond * 10 // a too small backoff will lead to high CPU load

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
