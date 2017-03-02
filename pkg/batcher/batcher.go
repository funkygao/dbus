// Package batcher provides retriable batch buffer: all succeed or rollback for all.
package batcher

import (
	"fmt"
	"sync/atomic"
	"time"
)

const backoff = time.Microsecond

// Batcher is a batched lockless queue that borrows design from disruptor.
//go:generate structlayout github.com/funkygao/dbus/pkg/batcher Batcher
// TODO padding
type Batcher struct {
	capacity   uint32
	stopped    uint32 // FIXME too much overhead
	w, r, c    uint32
	okN, failN uint32

	// [nil, item1, item2, ..., itemN]
	contents []interface{}
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

// String mainly used for debugging.
func (b *Batcher) String() string {
	return fmt.Sprintf("%+v", b.contents)
}

// TODO drain
func (b *Batcher) Close() {
	atomic.StoreUint32(&b.stopped, 1)
}

func (b *Batcher) Write(v interface{}) error {
	if atomic.LoadUint32(&b.stopped) == 1 {
		return ErrStopping
	}

	for atomic.LoadUint32(&b.w) == b.capacity+1 {
		// tail reached, wait for the current batch complete
		// FIXME what if Close called
		time.Sleep(backoff)
	}

	myIndex := atomic.AddUint32(&b.w, 1) - 1
	b.contents[myIndex] = v

	// mark the item available for reading
	atomic.StoreUint32(&b.c, myIndex)

	return nil
}

func (b *Batcher) ReadOne() (interface{}, error) {
	for {
		if atomic.LoadUint32(&b.stopped) == 1 {
			return nil, ErrStopping
		}

		r, w := atomic.LoadUint32(&b.r), atomic.LoadUint32(&b.w)
		if r == b.capacity+1 ||
			// tail reached, but not committed
			r >= w {
			// reader out-run writer
			time.Sleep(backoff)
		} else {
			break
		}
	}

	myIndex := atomic.AddUint32(&b.r, 1) - 1
	for myIndex > atomic.LoadUint32(&b.c) {
		time.Sleep(backoff)
	}

	return b.contents[int(myIndex)], nil
}

func (b *Batcher) Succeed() {
	ok := atomic.AddUint32(&b.okN, 1)
	if ok == b.capacity {
		// batch commit
		atomic.StoreUint32(&b.okN, 0)
		atomic.StoreUint32(&b.failN, 0)
		atomic.StoreUint32(&b.c, 0)
		atomic.StoreUint32(&b.w, 1)
		atomic.StoreUint32(&b.r, 1)
	} else if ok+atomic.LoadUint32(&b.failN) == b.capacity {
		// batch rollback, reset reader cursor, retry the batch, hold writer
		atomic.StoreUint32(&b.okN, 0)
		atomic.StoreUint32(&b.failN, 0)
		atomic.StoreUint32(&b.r, 1)
	}

}

func (b *Batcher) Fail() {
	fail := atomic.AddUint32(&b.failN, 1)
	if fail+atomic.LoadUint32(&b.okN) == b.capacity {
		// batch rollback
		atomic.StoreUint32(&b.okN, 0)
		atomic.StoreUint32(&b.failN, 0)
		atomic.StoreUint32(&b.r, 1)
	}
}
