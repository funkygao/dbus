package myslave

import (
	"context"
	"time"

	log "github.com/funkygao/log4go"
	"github.com/siddontang/go-mysql/mysql"
	"github.com/siddontang/go-mysql/replication"
)

// TODO graceful shutdown
// TODO GTID
func (m *MySlave) StartReplication(ready chan struct{}) {
	m.rowsEvent = make(chan *RowsEvent, m.c.Int("event_buffer_len", 100))
	m.errors = make(chan error, 1)

	m.r = replication.NewBinlogSyncer(&replication.BinlogSyncerConfig{
		ServerID: uint32(m.c.Int("server_id", 1007)),
		Flavor:   m.c.String("flavor", "mysql"),
		Host:     m.host,
		Port:     m.port,
		User:     m.c.String("user", "root"),
		Password: m.c.String("password", ""),
	})

	file, offset, err := m.p.Committed()
	if err != nil {
		log.Error("%s", err)
		m.emitFatalError(err)

		close(ready)
		return
	}

	syncer, err := m.r.StartSync(mysql.Position{
		Name: file,
		Pos:  offset,
	})
	if err != nil {
		// e,g.
		// ERROR 1045 (28000): Access denied for user 'xx'@'1.1.1.1'
		log.Error("%s", err)
		close(ready)
		m.emitFatalError(err)

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

		m.m.Events.Mark(1)

		log.Debug("-> %T", ev.Event)
		switch e := ev.Event.(type) {
		case *replication.RotateEvent:
			// e,g.
			// Position: 4
			// Next log name: mysql.000002
			file = string(e.NextLogName)
			log.Trace("rotate to (%s, %d)", file, e.Position)

		case *replication.RowsEvent:
			m.m.TPS.Mark(1)
			m.m.Lag.Update(time.Now().Unix() - int64(ev.Header.Timestamp))
			m.handleRowsEvent(file, ev.Header, e)

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
	if m.r != nil {
		m.r.Close()
	}
}
