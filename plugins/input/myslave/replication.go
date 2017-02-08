package myslave

import (
	"context"
	"fmt"
	"time"

	log "github.com/funkygao/log4go"
	"github.com/siddontang/go-mysql/replication"
)

// TODO graceful shutdown
// TODO GTID
func (m *MySlave) StartReplication(ready chan struct{}) {
	host := m.cf.String("master_host", "localhost")
	port := uint16(m.cf.Int("master_port", 3306))
	m.masterAddr = fmt.Sprintf("%s:%d", host, port)
	m.rowsEvent = make(chan *RowsEvent, m.cf.Int("event_buffer_len", 100))
	m.errors = make(chan error)

	m.r = replication.NewBinlogSyncer(&replication.BinlogSyncerConfig{
		ServerID: uint32(m.cf.Int("server_id", 1007)),
		Flavor:   m.cf.String("flavor", "mysql"),
		Host:     host,
		Port:     port,
		User:     m.cf.String("user", "root"),
		Password: m.cf.String("password", ""),
	})
	syncer, err := m.r.StartSync(m.pos.Pos())
	if err != nil {
		log.Error("%s", err)
		m.emitFatalError(err)

		close(ready)
		return
	}

	close(ready)
	log.Trace("ready to receive binlog stream")

	timeout := time.Second
	maxTimeout := time.Minute
	for {
		// what if the conn broken?
		// 2017/02/08 08:31:39 binlogsyncer.go:529: [info] receive EOF packet, retry ReadPacket
		// 2017/02/08 08:31:39 binlogsyncer.go:484: [error] connection was bad
		// 2017/02/08 08:31:40 binlogsyncer.go:446: [info] begin to re-sync from (mysql.000003, 2492)
		// 2017/02/08 08:31:40 binlogsyncer.go:130: [info] register slave for master server localhost:3306
		// 2017/02/08 08:31:40 binlogsyncer.go:502: [error] retry sync err: dial tcp [::1]:3306: getsockopt: connection refused, wait 1s and retry again
		// 2017/02/08 08:31:42 binlogsyncer.go:446: [info] begin to re-sync from (mysql.000003, 2492)
		// 2017/02/08 08:31:42 binlogsyncer.go:130: [info] register slave for master server localhost:3306
		// 2017/02/08 08:31:42 binlogsyncer.go:502: [error] retry sync err: dial tcp [::1]:3306: getsockopt: connection refused, wait 1s and retry again
		//
		// or invalid position sent?
		// 2017/02/08 08:42:09 binlogstreamer.go:47: [error] close sync with err: ERROR 1236 (HY000): Client requested master to start replication from position > file size; the first event 'mysql.000004' at 2274, the last event read from '/usr/local/var/mysql.000004' at 4, the last byte read from '/usr/local/var/mysql.000004' at 4.
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		ev, err := syncer.GetEvent(ctx)
		cancel()

		if err != nil {
			if err == context.DeadlineExceeded {
				if timeout < maxTimeout {
					// backoff
					timeout *= 2
				}

				continue
			}

			// TODO try out all the cases
			m.emitFatalError(err)
			return
		}

		log.Debug("-> %T", ev.Event)
		switch e := ev.Event.(type) {
		case *replication.RotateEvent:
			// e,g.
			// Position: 4
			// Next log name: mysql.000002
			m.pos.Name = string(e.NextLogName)
			// FIXME
			log.Trace("%s %d", m.pos.Name, e.Position)

		case *replication.RowsEvent:
			m.handleRowsEvent(ev.Header, e)

		case *replication.QueryEvent:
			// DDL comes this way
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

		case *replication.GTIDEvent:
			// TODO

		default:
			log.Warn("unexpected event: %+v", e)
		}
	}

}

func (m *MySlave) emitFatalError(err error) {
	m.errors <- err
	m.r.Close()
}
