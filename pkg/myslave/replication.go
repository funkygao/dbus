package myslave

import (
	"context"
	"fmt"
	"time"

	"github.com/funkygao/dbus/pkg/checkpoint"
	"github.com/funkygao/dbus/pkg/model"
	log "github.com/funkygao/log4go"
	uuid "github.com/satori/go.uuid"
	"github.com/siddontang/go-mysql/mysql"
	"github.com/siddontang/go-mysql/replication"
)

// StopReplication stops the slave and do necessary cleanups.
func (m *MySlave) StopReplication() {
	if !m.started.Get() {
		return
	}

	log.Debug("[%s] stopping replication from %s", m.name, m.masterAddr)

	m.r.Close()
	m.m.Close()

	if err := m.p.Shutdown(); err != nil {
		log.Error("[%s] %s", m.name, err)
	}

	m.started.Set(false)
}

// StartReplication start the mysql binlog replication.
// TODO graceful shutdown
// TODO GTID
func (m *MySlave) StartReplication(ready chan struct{}) {
	m.started.Set(true)

	m.rowsEvent = make(chan *model.RowsEvent, m.c.Int("event_buffer_len", 100))
	m.errors = make(chan error, 1)

	m.r = replication.NewBinlogSyncer(&replication.BinlogSyncerConfig{
		ServerID:        uint32(m.c.Int("server_id", 137)), // 137 unique enough? TODO
		Flavor:          m.c.String("flavor", mysql.MySQLFlavor),
		Host:            m.host,
		Port:            m.port,
		User:            m.user,
		Password:        m.passwd,
		SemiSyncEnabled: m.c.Bool("semi_sync", false),
	})

	// resume replication position from the checkpoint
	err := m.p.LastPersistedState(m.state)
	if err != nil && err != checkpoint.ErrStateNotFound {
		close(ready)
		m.emitFatalError(err)
		return
	}

	file, offset := m.state.File, m.state.Offset

	var syncer *replication.BinlogStreamer
	if m.GTID {
		// SELECT @@gtid_mode
		// SHOW GLOBAL VARIABLES LIKE 'SERVER_UUID'
		var masterUUID uuid.UUID
		// 07c93cd7-a7d3-12a5-94e1-a0369a7c3790:1-313225133
		set, _ := mysql.ParseMysqlGTIDSet(fmt.Sprintf("%s:%d-%d", masterUUID.String(), 1, 2))
		syncer, err = m.r.StartSyncGTID(set)
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
	log.Debug("[%s] ready to receive mysql binlog from %s", m.name, m.masterAddr)

	timeout := time.Second
	maxTimeout := time.Minute
	for m.started.Get() {
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
		// mysql> SHOW BINLOG EVENTS IN '8623306-bin.008517' FROM 117752230 limit 10;
		// +--------------------+-----------+------------+-----------+-------------+-----------------------------------------------------------------------------+
		// | Log_name           | Pos       | Event_type | Server_id | End_log_pos | Info                                                                        |
		// +--------------------+-----------+------------+-----------+-------------+-----------------------------------------------------------------------------+
		// | 8623306-bin.008517 | 117732471 | Query      |  81413306 |   117732534 | BEGIN                                                                       |
		// | 8623306-bin.008517 | 117732534 | Table_map  |  81413306 |   117732590 | table_id: 2925 (zabbix.history)                                             |
		// | 8623306-bin.008517 | 117732590 | Write_rows |  81413306 |   117735750 | table_id: 2925 flags: STMT_END_F                                            |
		// | 8623306-bin.008517 | 117735750 | Table_map  |  81413306 |   117735810 | table_id: 2968 (zabbix.history_uint)                                        |
		// | 8623306-bin.008517 | 117735810 | Write_rows |  81413306 |   117744020 | table_id: 2968                                                              |
		// | 8623306-bin.008517 | 117744020 | Write_rows |  81413306 |   117752230 | table_id: 2968                                                              |
		// | 8623306-bin.008517 | 117752230 | Write_rows |  81413306 |   117755340 | table_id: 2968 flags: STMT_END_F                                            | <- checkpoint
		// | 8623306-bin.008517 | 117755340 | Xid        |  81413306 |   117755371 | COMMIT /* xid=2243095633 */                                                 |
		// | 8623306-bin.008517 | 117755371 | Gtid       |  81413306 |   117755419 | SET @@SESSION.GTID_NEXT= 'a0f36bb1-fdbb-11e5-8413-a0369f7c3bb4:16393135780' |
		// | 8623306-bin.008517 | 117755419 | Query      |  81413306 |   117755482 | BEGIN                                                                       |
		// | 8623306-bin.008517 | 117755482 | Table_map  |  81413306 |   117755538 | table_id: 2925 (zabbix.history)                                             |
		// | 8623306-bin.008517 | 117755538 | Write_rows |  81413306 |   117758123 | table_id: 2925 flags: STMT_END_F                                            |
		// | 8623306-bin.008517 | 117758123 | Table_map  |  81413306 |   117758183 | table_id: 2968 (zabbix.history_uint)                                        |
		// | 8623306-bin.008517 | 117758183 | Write_rows |  81413306 |   117766393 | table_id: 2968                                                              |
		// | 8623306-bin.008517 | 117766393 | Write_rows |  81413306 |   117772728 | table_id: 2968 flags: STMT_END_F                                            |
		// | 8623306-bin.008517 | 117772728 | Xid        |  81413306 |   117772759 | COMMIT /* xid=2243095636 */                                                 |
		// +--------------------+-----------+------------+-----------+-------------+-----------------------------------------------------------------------------+
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		ev, err := syncer.GetEvent(ctx)
		//ev, err := syncer.GetEvent(context.Background())
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
			// e.Position is End_log_pos(i,e. next log position)
			log.Trace("[%s] events rotate to (%s, %d)", m.name, file, e.Position)

		case *replication.RowsEvent:
			m.m.TPS.Mark(1)
			// ev.Header.ServerID, where the event originated
			// depends on ntp to sync clock, but there is sure clock drift
			lag := time.Now().Unix() - int64(ev.Header.Timestamp)
			if lag < 0 {
				lag = 0
			}
			m.m.Lag.Update(lag)
			m.handleRowsEvent(file, ev.Header, e)

		case *replication.QueryEvent:
			// DDL comes this way
			// e,g. create table y(id int)
			// e,g. BEGIN
			// e,g. flush tables
			// e,g. ALTER TABLE
			//
			// only handles alter table query
			if db, table, yes := isAlterTableQuery(e.Query); yes {
				if len(db) == 0 {
					db = string(e.Schema)
				}
				log.Trace("[%s] %s.%s table schema changed", m.name, db, table)
				m.clearTableCache(db, table)
			}

		case *replication.XIDEvent:
			// e,g. COMMIT /* xid=403013040 */

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
			//
			// https://github.com/siddontang/go-mysql/issues/48
			// If we have processed TableMapEvent and then restart, the following
			// rows event may miss the table id when we sync from last saved position.

		case *replication.GenericEvent:
			// known misc types of event

		case *replication.GTIDEvent:
			// only GTID mode will recv this event

		default:
			log.Warn("[%s] unexpected event: %+v", m.name, e)
		}
	}

}

func (m *MySlave) emitFatalError(err error) {
	m.errors <- err
}
