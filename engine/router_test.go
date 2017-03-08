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

func BenchmarkRouterStatsUpdate(b *testing.B) {
	pack := NewPacket(nil)
	pack.input = true
	pack.Payload = Bytes("hello world")
	rs := routerStats{}
	for i := 0; i < b.N; i++ {
		rs.update(pack)
	}
}
