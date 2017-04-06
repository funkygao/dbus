package batcher

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

func TestGolangPadding(t *testing.T) {
	type Compact struct {
		a, b                   uint64
		c, d, e, f, g, h, i, j byte
	}

	// Larger memory footprint than "Compact" - but less fields!
	// The values are padded to 8 bytes on x64
	type Inefficient struct {
		a uint64
		b byte
		c uint64
		d byte
	}

	newCompact := new(Compact)
	t.Log(unsafe.Sizeof(*newCompact)) // 24
	newInefficient := new(Inefficient)
	t.Log(unsafe.Sizeof(*newInefficient)) // 32
}

func TestBatcherBasic(t *testing.T) {
	b := New(8)
	// inject [0, 8)
	for i := 0; i < 8; i++ {
		b.Put(i)
	}
	t.Logf("[0, 8) injected, %+v", b)

	// [0, 1] ok
	for i := 0; i < 2; i++ {
		b.Succeed()
	}
	t.Logf("[0, 1] ok, %+v", b)

	go func() {
		for {
			v, err := b.Get()
			if err != nil {
				return
			}

			t.Logf("ReadOne<- %v, %+v", v, b)
		}
	}()

	// [3, ...] fails
	for i := 2; i < 8; i++ {
		b.Fail()
		t.Logf("[%d] fail, %+v", i, b)
	}

	//b.Write(5) // will block

	t.Logf("sleep 1s for reader catch up")
	time.Sleep(time.Second)
	t.Logf("closing batcher")
	b.Close()

}

func BenchmarkBatcherGetWithBatchSize100(b *testing.B) {
	batcher := New(100)
	go func() {
		for {
			batcher.Put(1)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := batcher.Get()
		if err != nil {
			panic(err)
		}
		batcher.Succeed()
	}
}

func BenchmarkBatcherPutWithBatchSize100(b *testing.B) {
	batcher := New(100)
	go func() {
		for {
			_, err := batcher.Get()
			if err != nil {
				panic(err)
			}
			batcher.Succeed()
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batcher.Put(1)
	}
}

func BenchmarkBatcherPutWithBatchSize1000(b *testing.B) {
	batcher := New(1000)
	go func() {
		for {
			_, err := batcher.Get()
			if err != nil {
				panic(err)
			}
			batcher.Succeed()
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batcher.Put(1)
	}
}

func BenchmarkWaitGroup(b *testing.B) {
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		wg.Done()
		wg.Wait()
	}
}

func BenchmarkAtomicAdd(b *testing.B) {
	var v uint64
	for i := 0; i < b.N; i++ {
		atomic.AddUint64(&v, 1)
	}
}

func BenchmarkAtomicCAS(b *testing.B) {
	var v uint64
	for i := 0; i < b.N; i++ {
		atomic.CompareAndSwapUint64(&v, 0, 1)
	}
}
