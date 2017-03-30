package binlog

import (
	"encoding/json"

	"github.com/funkygao/dbus/pkg/checkpoint"
)

var (
	_ checkpoint.State = &BinlogState{}
)

type BinlogState struct {
	File   string `json:"file"`
	Offset uint32 `json:"offset"`
}

func New() *BinlogState {
	return &BinlogState{}
}

func (s *BinlogState) Marshal() []byte {
	b, _ := json.Marshal(s)
	return b
}

func (s *BinlogState) Unmarshal(data []byte) {
	json.Unmarshal(data, s)
}

func (s *BinlogState) Reset() {
	s.File = ""
	s.Offset = 0
}
