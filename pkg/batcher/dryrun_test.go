package batcher

import (
	"testing"
	"time"
)

func TestDryrunBasic(t *testing.T) {
	b := NewDryrun(8)
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
