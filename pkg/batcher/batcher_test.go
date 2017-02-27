package batcher

import (
	"testing"
	"time"
)

func TestBatcherBasic(t *testing.T) {
	b := NewBatcher(8)
	// inject [0, 8)
	for i := 0; i < 8; i++ {
		b.Write(i)
	}
	t.Logf("[0, 8) injected, %+v", b)

	// [0, 1] ok
	for i := 0; i < 2; i++ {
		b.Advance()
	}
	t.Logf("[0, 1] ok, %+v", b)

	go func() {
		for {
			v, err := b.ReadOne()
			if err != nil {
				return
			}

			t.Logf("ReadOne<- %v, %+v", v, b)
		}
	}()

	// [3, ...] fails
	for i := 2; i < 8; i++ {
		b.Rollback()
		t.Logf("[%d] fail, %+v", i, b)
	}

	//b.Write(5) // will block

	t.Logf("sleep 1s for reader catch up")
	time.Sleep(time.Second)
	t.Logf("closeing batcher")
	b.Close()

}

func BenchmarkBatcherReadWrite1kInBatch(b *testing.B) {
	batcher := NewBatcher(1 << 10)
	go func() {
		for {
			batcher.Write(1)
		}
	}()

	for i := 0; i < b.N; i++ {
		_, err := batcher.ReadOne()
		if err != nil {
			panic(err)
		}
		batcher.Advance()
	}
}
