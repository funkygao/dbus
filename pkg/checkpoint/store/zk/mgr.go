package zk

import (
	"github.com/funkygao/dbus/pkg/checkpoint"
	"github.com/funkygao/gafka/zk"
)

var _ checkpoint.Manager = &manager{}

type manager struct {
	zkzone *zk.ZkZone
}

func NewManager(zkzone *zk.ZkZone) checkpoint.Manager {
	return &manager{zkzone: zkzone}
}

func (m *manager) AllStates() ([]checkpoint.State, error) {
	return nil, nil
}
