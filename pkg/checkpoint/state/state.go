package state

import (
	"errors"
	"net/url"

	"github.com/funkygao/dbus/pkg/checkpoint"
	"github.com/funkygao/dbus/pkg/checkpoint/state/binlog"
	"github.com/funkygao/dbus/pkg/checkpoint/state/kafka"
)

func From(scheme string, rawDSN, name string, data []byte) (checkpoint.State, error) {
	dsn, err := url.QueryUnescape(rawDSN)
	if err != nil {
		return nil, err
	}

	var s checkpoint.State
	switch scheme {
	case checkpoint.SchemeKafka:
		s = kafka.New(dsn, name)

	case checkpoint.SchemeBinlog:
		s = binlog.New(dsn, name)

	default:
		return nil, errors.New("invalid scheme")
	}

	s.Unmarshal(data)
	return s, nil
}
