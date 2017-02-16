package myslave

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/funkygao/dbus/pkg/model"
	log "github.com/funkygao/log4go"
	uuid "github.com/satori/go.uuid"
	"github.com/siddontang/go-mysql/mysql"
	"github.com/siddontang/go-mysql/replication"
)

func (m *MySlave) StopReplication() {
	if m.isMaster {
		m.r.Close()
		if err := m.p.Flush(); err != nil {
			log.Error("[%s] flush: %s", m.name, err)
		}
	}

	m.leaveCluster()
	m.isMaster = false
}

// TODO graceful shutdown
// TODO GTID
func (m *MySlave) StartReplication(ready chan struct{}) {
	m.joinClusterAndBecomeMaster() // block till become master
	m.isMaster = true

	m.rowsEvent = make(chan *model.RowsEvent, m.c.Int("event_buffer_len", 100))
	m.errors = make(chan error, 1)

	m.r = replication.NewBinlogSyncer(&replication.BinlogSyncerConfig{
		ServerID:        uint32(m.c.Int("server_id", 137)), // 137 unique enough?
		Flavor:          "mysql",                           // currently mariadb not implemented
		Host:            m.host,
		Port:            m.port,
		User:            m.c.String("user", "root"),
		Password:        m.c.String("password", ""),
		SemiSyncEnabled: false,
	})

	file, offset, err := m.p.Committed()
	if err != nil {
		close(ready)
		m.emitFatalError(err)
		return
	}

	var syncer *replication.BinlogStreamer
	if m.GTID {
		// SELECT @@gtid_mode
		// SHOW GLOBAL VARIABLES LIKE 'SERVER_UUID'
		var masterUuid uuid.UUID
		// 07c93cd7-a7d3-12a5-94e1-a0369a7c3790:1-313225133
		set, _ := mysql.ParseMysqlGTIDSet(fmt.Sprintf("%s:%d-%d", masterUuid.String(), 1, 2))
		syncer, err = m.r.StartSyncGTID(set)
		panic("not implemented") // TODO
	} else {
		syncer, err = m.r.StartSync(mysql.Position{
			Name: file,
			Pos:  offset,
		})
	}

	if err != nil {
		// unrecoverable err encountered
		// e,g.
		// ERROR 1045 (28000): Access denied for user 'xx'@'1.1.1.1'
		close(ready)
		m.emitFatalError(err)
		return
	}

	close(ready)
	log.Trace("[%s] ready to receive binlog stream", m.name)

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
		//
		// [02/08/17 17:39:20] [EROR] ( mysqlbinlog.go:64) backoff 5s: invalid table id 2413, no correspond table map event
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

			if strings.HasPrefix(err.Error(), "invalid table id") {
				log.Error("[%s] %s", m.name, err)
				// TODO how to handle this?
				continue
			}

			// TODO try out all the cases
			m.emitFatalError(err)
			return
		}

		m.m.Events.Mark(1)

		log.Debug("-> %T", ev.Event)
		// insert into tbtest values(1) will trigger the following events:
		// QueryEvent    BEGIN, Log position: 4800
		// TableMapEvent Schema: mydb, Table: tbtest, TableID: 81, Log position: 4844
		// RowsEvent     Values: 0:1, TableID: 81, Log position: 4884
		// XIDEvent      XID: 356, Log position: 4915 // COMMMIT
		switch e := ev.Event.(type) {
		case *replication.RotateEvent:
			// e,g.
			// Position: 4
			// Next log name: mysql.000002
			file = string(e.NextLogName)
			log.Trace("[%s] rotate to (%s, %d)", m.name, file, e.Position)

		case *replication.RowsEvent:
			m.m.TPS.Mark(1)
			m.m.Lag.Update(time.Now().Unix() - int64(ev.Header.Timestamp))
			m.handleRowsEvent(file, ev.Header, e)

		case *replication.QueryEvent:
			// DDL comes this way
			// e,g. create table y(id int)
			// e,g. BEGIN
			// e,g. flush tables

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
			// known misc types of event

		case *replication.GTIDEvent:
			if m.GTID {
				// TODO
			}

		default:
			log.Warn("[%s] unexpected event: %+v", m.name, e)
		}
	}

}

func (m *MySlave) emitFatalError(err error) {
	m.errors <- err
	if m.r != nil {
		m.r.Close()
	}
}
