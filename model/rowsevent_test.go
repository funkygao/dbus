package model

import (
	"testing"
)

func makeRowsEvent() *RowsEvent {
	return &RowsEvent{
		Log:       "mysql-bin.0001",
		Position:  498876,
		Schema:    "mydabase",
		Table:     "user_account",
		Action:    "I",
		Timestamp: 1486554654,
		Rows:      [][]interface{}{{"user", 15, "hello world"}},
	}
}

func BenchmarkRowsEventEncode(b *testing.B) {
	r := makeRowsEvent()
	for i := 0; i < b.N; i++ {
		r.Encode()
	}
}

func BenchmarkRowsEventLength(b *testing.B) {
	r := makeRowsEvent()
	for i := 0; i < b.N; i++ {
		r.Length()
	}
}
