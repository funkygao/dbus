package engine

import (
	"testing"
)

func BenchmarkPack(b *testing.B) {
	pool := 100
	inChan := make(chan *PipelinePack, pool)
	for i := 0; i < pool; i++ {
		pack := NewPipelinePack(inChan)
		inChan <- pack
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := <-inChan
		p.Recycle()
	}
}
