package kafka

import (
	"encoding/json"
	"fmt"

	"github.com/funkygao/dbus/pkg/checkpoint"
)

var (
	_ checkpoint.State = &KafkaState{}
)

type KafkaState struct {
	dsn  string
	name string // who updates this state

	PartitionID int32 `json:"pid"`
	Offset      int64 `json:"offset"`
}

// New creates a kafka state.
// dsn is the DSN of kafka conumer, name is the Input plugin name.
func New(dsn string, name string) *KafkaState {
	return &KafkaState{dsn: dsn, name: name}
}

func (s *KafkaState) Marshal() []byte {
	b, _ := json.Marshal(s)
	return b
}

func (s *KafkaState) Unmarshal(data []byte) {
	json.Unmarshal(data, s)
}

func (s *KafkaState) reset() {
	s.PartitionID = 0
	s.Offset = 0
}

func (s *KafkaState) Name() string {
	return s.name
}

func (s *KafkaState) String() string {
	return fmt.Sprintf("%d-%d", s.PartitionID, s.Offset)
}

func (s *KafkaState) DSN() string {
	return s.dsn
}

func (s *KafkaState) Scheme() string {
	return checkpoint.SchemeKafka
}
