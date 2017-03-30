package binlog

import (
	"encoding/json"

	"github.com/funkygao/dbus/pkg/checkpoint"
)

var (
	_ checkpoint.State = &state{}
)

type state struct {
	File   string `json:"file"`
	Offset uint32 `json:"offset"`
}

func New() *state {
	return &state{}
}

func (s *state) Marshal() []byte {
	b, _ := json.Marshal(s)
	return b
}

func (s *state) Unmarshal(data []byte) {
	json.Unmarshal(data, s)
}

func (s *state) Reset() {
	s.File = ""
	s.Offset = 0
}
