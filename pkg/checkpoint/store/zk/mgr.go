package zk

import (
	"path"

	"github.com/funkygao/dbus/pkg/checkpoint"
	"github.com/funkygao/dbus/pkg/checkpoint/state/binlog"
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
	schemes, _, err := m.zkzone.Conn().Children(root)
	if err != nil {
		return nil, err
	}

	var r []checkpoint.State
	for _, scheme := range schemes {
		dsns, _, err := m.zkzone.Conn().Children(path.Join(root, scheme))
		if err != nil {
			return nil, err
		}

		for _, dsn := range dsns {
			data, _, err := m.zkzone.Conn().Get(path.Join(root, scheme, dsn))
			if err != nil {
				return nil, err
			}

			// FIXME ugly design
			switch scheme {
			case "myslave":
				s := binlog.New(dsn, "") // empty name is ok, wait for Unmarshal
				s.Unmarshal(data)
				r = append(r, s)

			case "kafka":
				// TODO

			default:
				panic("unknown scheme: " + scheme)
			}
		}
	}

	return r, nil
}
