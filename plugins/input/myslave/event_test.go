package myslave

import (
	"testing"
)

func BenchmarkRowEventMarshal(b *testing.B) {
	r := &RowsEvent{
		Log:       "mysql-bin.0001",
		Position:  498876,
		Schema:    "mydabase",
		Table:     "user_account",
		Action:    "I",
		Timestamp: 1486554654,
		Rows:      [][]interface{}{{"user", 15, "hello world"}},
	}

	for i := 0; i < b.N; i++ {
		r.Bytes()
	}
}
