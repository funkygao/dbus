package engine

import (
	"runtime"
	"sync/atomic"
)

const (
	queueSize uint64 = 4096
	indexMask uint64 = queueSize - 1
)

// Channel is a lock-free ringbuffer that is 2X faster than golang channel.
// However, it is lacking in the rich features of golang channel.
type Channel struct {
	_padding1          [8]uint64
	lastCommittedIndex uint64

	_padding2     [8]uint64
	nextFreeIndex uint64

	_padding3   [8]uint64
	readerIndex uint64

	_padding4 [8]uint64
	contents  [queueSize]interface{}

	_padding5 [8]uint64
}

func NewChannel() *Channel {
	return &Channel{lastCommittedIndex: 0, nextFreeIndex: 1, readerIndex: 1}
}

func (ch *Channel) Put(value interface{}) {
	var myIndex = atomic.AddUint64(&ch.nextFreeIndex, 1) - 1
	for myIndex > (ch.readerIndex + queueSize - 2) {
		runtime.Gosched()
	}

	ch.contents[myIndex&indexMask] = value

	for !atomic.CompareAndSwapUint64(&ch.lastCommittedIndex, myIndex-1, myIndex) {
		runtime.Gosched()
	}
}

func (ch *Channel) Get() interface{} {
	var myIndex = atomic.AddUint64(&ch.readerIndex, 1) - 1
	for myIndex > ch.lastCommittedIndex {
		runtime.Gosched()
	}

	return ch.contents[myIndex&indexMask]
}
