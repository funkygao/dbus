package engine

import (
	"testing"
)

type Bytes []byte

func (b Bytes) Length() int {
	return len(b)
}

func (b Bytes) String() string {
	return string(b)
}

func (b Bytes) Encode() ([]byte, error) {
	return b, nil
}

func BenchmarkRouterMetrics(b *testing.B) {
	pack := newPacket(nil)
	pack.Payload = Bytes("hello world")
	for i := 0; i < b.N; i++ {
		_ = pack
	}
}
