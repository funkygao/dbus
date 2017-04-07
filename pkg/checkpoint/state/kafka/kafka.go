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
	dsn string

	PartitionID int32
	Offset      int64
}

func New(dsn string) *KafkaState {
	return &KafkaState{dsn: dsn}
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

func (s *KafkaState) String() string {
	return fmt.Sprintf("%d-%d", s.PartitionID, s.Offset)
}

func (s *KafkaState) DSN() string {
	return s.dsn
}

func (s *KafkaState) Scheme() string {
	return "kafka"
}
