package engine

import (
	"testing"
)

func BenchmarkPackRecycle(b *testing.B) {
	poolSize := 100
	inChan := make(chan *Packet, poolSize)
	for i := 0; i < poolSize; i++ {
		pack := NewPacket(inChan)
		inChan <- pack
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := <-inChan
		p.Recycle()
	}
}

func BenchmarkMapGetItem(b *testing.B) {
	m := map[string]struct{}{
		"hello": struct{}{},
	}

	for i := 0; i < b.N; i++ {
		_ = m["hello"]
	}
}
