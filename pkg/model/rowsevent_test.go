package model

import (
	"encoding/json"
	"testing"

	"github.com/funkygao/assert"
	"gopkg.in/vmihailenco/msgpack.v2"
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

func TestAvroMarshaller(t *testing.T) {
	t.SkipNow()

	r := makeRowsEvent()
	b, err := avroMarshaller(r)
	assert.Equal(t, nil, err)
	assert.Equal(t, "exp", string(b))
}

func TestRowsEventFlags(t *testing.T) {
	r := makeRowsEvent()
	r.SetFlags(1)
	assert.Equal(t, true, r.IsStmtEnd())
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
	for i := 0; i < b.N; i++ {
		r := makeRowsEvent()
		json.Marshal(r)
	}
}

func BenchmarkMsgpackEncodeRowsEvent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := makeRowsEvent()
		msgpack.Marshal(r)
	}
}

func BenchmarkRowsEventJsonMarshalFF(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := makeRowsEvent()
		r.MarshalJSON()
	}
}
