package myslave

import (
	"context"
	"time"

	log "github.com/funkygao/log4go"
	"github.com/siddontang/go-mysql/replication"
)

func (m *MySlave) StartReplication(ready chan struct{}) {
	m.rowsEvent = make(chan *RowsEvent, 100)
	m.errors = make(chan error)

	if m.gtid {
		// TODO
	}

	s, err := m.r.StartSync(m.pos.Pos())
	if err != nil {
		m.errors <- err
		close(ready)
		return
	}

	close(ready)
	timeout := time.Second
	for {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		ev, err := s.GetEvent(ctx)
		cancel()

		if err != nil {
			if err == context.DeadlineExceeded {
				if timeout < time.Minute {
					// backoff
					timeout *= 2
				}

				continue
			}

			m.errors <- err
			return
		}

		log.Debug("-> %T", ev.Event)
		switch e := ev.Event.(type) {
		case *replication.RotateEvent:
			// e,g.
			// Position: 4
			// Next log name: mysql.000002
			m.pos.Name = string(e.NextLogName)

		case *replication.RowsEvent:
			m.handleRowsEvent(ev.Header, e)

		case *replication.QueryEvent:
			// e,g. create table y(id int)
			// e,g. BEGIN

		case *replication.XIDEvent:
			// e,g. xid: 1293

		case *replication.FormatDescriptionEvent:
			// Version: 4
			// Server version: 5.6.23-log
			// Checksum algorithm: 1

		case *replication.TableMapEvent:
			// e,g.
			// TableID: 170
			// TableID size: 6
			// Flags: 1
			// Schema: test
			// Table: y
			// Column count: 1
			// Column type:
			// 00000000  03
			// NULL bitmap:
			// 00000000  01

		case *replication.GenericEvent:

		default:
			log.Warn("unexpected event: %+v", e)

		}
	}

}
