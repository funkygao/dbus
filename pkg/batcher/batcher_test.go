package batcher

import (
	"testing"
	"time"
)

func TestBatcherBasic(t *testing.T) {
	b := NewBatcher(8)
	//
	for i := 0; i < 8; i++ {
		b.Write(i)
	}
	go func() {
		for {
			v := b.Read()
			t.Logf("%v", v)
		}
	}()
	t.Logf("%+v", b)
	// [0, 1] ok
	for i := 0; i < 2; i++ {
		b.Advance()
	}
	t.Logf("%+v", b)
	// [3, ...] fails
	b.Rollback()
	t.Logf("%+v", b)

	time.Sleep(time.Second)
}
