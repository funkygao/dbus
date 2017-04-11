package binlog

import (
	"encoding/json"
	"fmt"

	"github.com/funkygao/dbus/pkg/checkpoint"
)

var (
	_ checkpoint.State = &BinlogState{}
)

type BinlogState struct {
	dsn string

	Ident  string `json:"ident,omitempty"`
	File   string `json:"file"`
	Offset uint32 `json:"offset"`
}

func New(dsn string, name string) *BinlogState {
	return &BinlogState{dsn: dsn, Ident: name}
}

func (s *BinlogState) Marshal() []byte {
	b, _ := json.Marshal(s)
	return b
}

func (s *BinlogState) Unmarshal(data []byte) {
	json.Unmarshal(data, s)
}

func (s *BinlogState) reset() {
	s.File = ""
	s.Offset = 0
}

func (s *BinlogState) String() string {
	return fmt.Sprintf("%s-%d", s.File, s.Offset)
}

func (s *BinlogState) Name() string {
	return s.Ident
}

func (s *BinlogState) DSN() string {
	return s.dsn
}

func (s *BinlogState) Scheme() string {
	return "myslave"
}
