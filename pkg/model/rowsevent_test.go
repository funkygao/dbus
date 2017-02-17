package model

import (
	"encoding/json"
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

func TestRowsEventEncode(t *testing.T) {
	r := makeRowsEvent()
	b, err := r.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"log":"mysql-bin.0001","pos":498876,"db":"mydabase","tbl":"user_account","dml":"I","ts":1486554654,"rows":[["user",15,"hello world"]]}` {
		t.Fatal("encoded wrong:" + string(b))
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

func BenchmarkJsonEncodeRowsEvent(b *testing.B) {
	r := makeRowsEvent()
	for i := 0; i < b.N; i++ {
		json.Marshal(r)
	}
}
