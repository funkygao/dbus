// Package batcher provides retriable batch queue: all succeed or rollback for all.
// In kafka it is called RecordAccumulator.
package batcher

import (
	"sync/atomic"

	"github.com/funkygao/dbus/pkg/sys"
)

// Batcher is a batched lock free queue that borrows design from disruptor.
// It maintains a queue with the sematics of all succeed and advance or any fails and retry.
type Batcher struct {
	_padding0 [sys.CacheLineSize / 8]uint64
	capacity  uint32

	_padding1 [sys.CacheLineSize / 8]uint64
	stopped   uint32

	_padding2 [sys.CacheLineSize / 8]uint64
	w         uint32

	_padding3 [sys.CacheLineSize / 8]uint64
	r         uint32

	_padding4 [sys.CacheLineSize / 8]uint64
	c         uint32

	_padding5 [sys.CacheLineSize / 8]uint64
	okN       uint32

	_padding6 [sys.CacheLineSize / 8]uint64
	failN     uint32

	// [nil, item1, item2, ..., itemN]
	_padding8 [sys.CacheLineSize / 8]uint64
	contents  []interface{}
}

// NewBatcher create a new smart batcher instance.
func NewBatcher(capacity int) *Batcher {
	return &Batcher{
		capacity: uint32(capacity),
		stopped:  0,
		w:        1, // 0 is reserved for c, w ranges in [1, n+1]
		r:        1, // 0 is reserved for c, r ranges in [1, n+1]
		c:        0, // [0, n]
		contents: make([]interface{}, capacity+1),
	}
}

// Close closes the batcher from R/W.
func (b *Batcher) Close() {
	atomic.StoreUint32(&b.stopped, 1)
}

// Put writes an item to the batcher. If queue is full, it will block till
// all inflight items marked success.
func (b *Batcher) Put(item interface{}) error {
	for atomic.LoadUint32(&b.w) == b.capacity+1 {
		if atomic.LoadUint32(&b.stopped) == 1 {
			return ErrStopping
		}

		// tail reached, wait for the current batch successful completion
		yield()
	}

	myIndex := atomic.AddUint32(&b.w, 1) - 1
	b.contents[myIndex] = item

	// mark the item available for reading
	atomic.StoreUint32(&b.c, myIndex)
	return nil
}

// Get reads an item from the batcher.
func (b *Batcher) Get() (interface{}, error) {
	for {
		r, w := atomic.LoadUint32(&b.r), atomic.LoadUint32(&b.w)
		if r == b.capacity+1 ||
			// tail reached, but not committed
			r >= w {
			// reader out-run writer

			if atomic.LoadUint32(&b.stopped) == 1 {
				return nil, ErrStopping
			}

			yield()
		} else {
			break
		}
	}

	myIndex := atomic.AddUint32(&b.r, 1) - 1
	for myIndex > atomic.LoadUint32(&b.c) {
		if atomic.LoadUint32(&b.stopped) == 1 {
			return nil, ErrStopping
		}

		// reader out-run commit sequence
		yield()
	}

	return b.contents[int(myIndex)], nil
}

// Succeed marks an item handling success.
func (b *Batcher) Succeed() {
	if okN := atomic.AddUint32(&b.okN, 1); okN == b.capacity {
		b.commit()
	} else if okN+atomic.LoadUint32(&b.failN) == b.capacity {
		b.rewind()
	}
}

// Fail marks an item handling failure.
func (b *Batcher) Fail() (rewind bool) {
	if failN := atomic.AddUint32(&b.failN, 1); failN+atomic.LoadUint32(&b.okN) == b.capacity {
		b.rewind()
		rewind = true
	}

	return
}

func (b *Batcher) commit() {
	atomic.StoreUint32(&b.okN, 0)
	atomic.StoreUint32(&b.failN, 0)
	atomic.StoreUint32(&b.c, 0)
	atomic.StoreUint32(&b.w, 1)
	atomic.StoreUint32(&b.r, 1)
}

// batch rollback, reset reader cursor, retry the batch, hold writer
func (b *Batcher) rewind() {
	atomic.StoreUint32(&b.okN, 0)
	atomic.StoreUint32(&b.failN, 0)
	atomic.StoreUint32(&b.r, 1)
}
