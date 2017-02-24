// Package batcher provides retriable batch buffer: all succeed or rollback for all.
package batcher

import (
	"sync/atomic"
	"time"
)

type Batcher struct {
	size     uint32
	w, r     uint32
	okN      uint32
	contents []interface{}
}

func NewBatcher(size int) *Batcher {
	return &Batcher{
		size:     uint32(size),
		contents: make([]interface{}, size),
	}
}

func (b *Batcher) Write(v interface{}) {
	for atomic.LoadUint32(&b.w) == b.size {
		// wait for the complete batch commit
		time.Sleep(time.Millisecond * 50)
	}

	idx := atomic.AddUint32(&b.w, 1) - 1 // lock step
	b.contents[idx] = v
}

func (b *Batcher) Read() interface{} {
	for atomic.LoadUint32(&b.r) == b.size {
		// read to tail while wait for commit/rollback
		time.Sleep(time.Millisecond * 50)
	}

	idx := atomic.AddUint32(&b.r, 1) - 1
	return b.contents[int(idx)]
}

func (b *Batcher) Advance() {
	ok := atomic.AddUint32(&b.okN, 1)
	if ok == b.size {
		atomic.StoreUint32(&b.okN, 0)
		atomic.StoreUint32(&b.w, 0)
		atomic.StoreUint32(&b.r, 0)
	}
}

func (b *Batcher) Rollback() {
	atomic.StoreUint32(&b.r, 0)
}
