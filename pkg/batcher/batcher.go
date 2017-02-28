// Package batcher provides retriable batch buffer: all succeed or rollback for all.
package batcher

import (
	"log"
	"sync"
	"sync/atomic"
	"time"
)

const backoff = time.Millisecond

type Batcher struct {
	size       uint32
	w, r       uint32
	okN, failN uint32
	barrier    sync.WaitGroup
	stopped    uint32
	contents   []interface{}
}

func NewBatcher(size int) *Batcher {
	b := &Batcher{
		size:     uint32(size),
		stopped:  0,
		contents: make([]interface{}, size),
	}
	b.barrier.Add(1)
	return b
}

func (b *Batcher) Close() {
	atomic.StoreUint32(&b.stopped, 1)
}

func (b *Batcher) Write(v interface{}) error {
	if atomic.LoadUint32(&b.w) == b.size {
		// wait for the batch complete
		b.barrier.Wait()

		// start a new batch
		b.barrier.Add(1)
	}

	if atomic.LoadUint32(&b.stopped) == 1 {
		return ErrStopping
	}

	idx := atomic.AddUint32(&b.w, 1) - 1
	b.contents[idx] = v
	log.Printf("%d %+v", idx, v)
	return nil
}

func (b *Batcher) ReadOne() (interface{}, error) {
	for {
		if atomic.LoadUint32(&b.stopped) == 1 {
			return nil, ErrStopping
		}

		r, w := atomic.LoadUint32(&b.r), atomic.LoadUint32(&b.w)
		log.Printf("r=%d w=%d", r, w)
		if r == b.size ||
			// batch tail reached, but not committed
			r >= w {
			// reader out-run writer
			time.Sleep(backoff)
		} else {
			break
		}
	}

	idx := atomic.AddUint32(&b.r, 1) - 1
	return b.contents[int(idx)], nil
}

func (b *Batcher) Advance() {
	ok := atomic.AddUint32(&b.okN, 1)
	if ok == b.size {
		// batch commit
		atomic.StoreUint32(&b.okN, 0)
		atomic.StoreUint32(&b.failN, 0)
		atomic.StoreUint32(&b.w, 0)
		atomic.StoreUint32(&b.r, 0)
		b.barrier.Done()
	} else if ok+atomic.LoadUint32(&b.failN) == b.size {
		// batch rollback, reset reader cursor, retry the batch, hold writer
		atomic.StoreUint32(&b.okN, 0)
		atomic.StoreUint32(&b.failN, 0)
		atomic.StoreUint32(&b.r, 0)
	}

}

func (b *Batcher) Rollback() {
	fail := atomic.AddUint32(&b.failN, 1)
	if fail+atomic.LoadUint32(&b.okN) == b.size {
		// batch rollback
		atomic.StoreUint32(&b.okN, 0)
		atomic.StoreUint32(&b.failN, 0)
		atomic.StoreUint32(&b.r, 0)
	}
}
