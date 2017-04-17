package state

import (
	"errors"
	"net/url"

	"github.com/funkygao/dbus/pkg/checkpoint"
	"github.com/funkygao/dbus/pkg/checkpoint/state/binlog"
	"github.com/funkygao/dbus/pkg/checkpoint/state/kafka"
)

// Load creates a state from its serialized data.
func Load(scheme string, rawDSN string, data []byte) (checkpoint.State, error) {
	dsn, err := url.QueryUnescape(rawDSN)
	if err != nil {
		return nil, err
	}

	var s checkpoint.State
	switch scheme {
	case checkpoint.SchemeKafka:
		s = kafka.New(dsn, "")

	case checkpoint.SchemeBinlog:
		s = binlog.New(dsn, "")

	default:
		return nil, errors.New("invalid scheme")
	}

	s.Unmarshal(data)
	return s, nil
}
