package myslave

import (
	"fmt"
	"time"

	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/pkg/checkpoint"
	"github.com/funkygao/dbus/pkg/checkpoint/state/binlog"
	czk "github.com/funkygao/dbus/pkg/checkpoint/store/zk"
	"github.com/funkygao/dbus/pkg/model"
	"github.com/funkygao/gafka/zk"
	"github.com/funkygao/golib/sync2"
	conf "github.com/funkygao/jsconf"
	mylog "github.com/ngaut/log"
	"github.com/siddontang/go-mysql/client"
	"github.com/siddontang/go-mysql/replication"
)

// MySlave is a mimic mysql slave that replicates binlog from mysql master using IO thread.
type MySlave struct {
	c     *conf.Conf
	r     *replication.BinlogSyncer
	p     checkpoint.Checkpoint
	m     *slaveMetrics
	z     *zk.ZkZone
	conn  *client.Conn
	state *binlog.BinlogState

	name string

	Predicate func(schema, table string) bool
	GTID      bool // global tx id

	dsn          string
	masterAddr   string
	host         string
	port         uint16
	user, passwd string

	dbAllowed  map[string]struct{}
	dbExcluded map[string]struct{}

	started   sync2.AtomicBool
	errors    chan error
	rowsEvent chan *model.RowsEvent
}

// New creates a MySlave instance.
func New(dsn string) *MySlave {
	// github.com/siddontang/go-mysql is using github.com/ngaut/log
	mylog.SetLevel(mylog.LOG_LEVEL_ERROR)
	if err := mylog.SetOutputByName("myslave.log"); err != nil {
		panic(err)
	}

	return &MySlave{
		dsn:        dsn,
		dbExcluded: map[string]struct{}{},
		dbAllowed:  map[string]struct{}{},
		state:      binlog.New(dsn),
	}
}

// LoadConfig initialize internal state according to the config section.
func (m *MySlave) LoadConfig(config *conf.Conf) *MySlave {
	m.c = config

	var err error
	var dbs []string
	var zone string
	zone, m.host, m.port, m.user, m.passwd, dbs, err = ParseDSN(m.dsn)
	if m.user == "" || zone == "" || m.host == "" || m.port == 0 || err != nil {
		panic("invalid dsn")
	}
	m.masterAddr = fmt.Sprintf("%s:%d", m.host, m.port)

	for _, db := range dbs {
		m.dbAllowed[db] = struct{}{}
	}
	for _, db := range config.StringList("db_excluded", nil) {
		m.dbExcluded[db] = struct{}{}
	}
	if len(m.dbAllowed) > 0 && len(m.dbExcluded) > 0 {
		panic("db_excluded and db allowed cannot be set at the same time")
	}
	m.setupPredicate()

	m.name = m.c.String("name", m.masterAddr)
	m.GTID = m.c.Bool("GTID", false)
	if m.GTID {
		panic("GTID not implemented")
	}

	m.m = newMetrics(m.host, m.port)
	m.z = engine.Globals().GetOrRegisterZkzone(zone)
	m.p = czk.New(m.z, m.state, m.masterAddr, m.c.Duration("pos_commit_interval", time.Second))

	return m
}

// MarkAsProcessed notifies the positioner that a certain binlog event
// has been successfully processed and should be committed.
func (m *MySlave) MarkAsProcessed(r *model.RowsEvent) error {
	if !r.IsStmtEnd() {
		// +--------------------+-----------+------------+-----------+-------------+-----------------------------------------------------------------------------+
		// | Log_name           | Pos       | Event_type | Server_id | End_log_pos | Info                                                                        |
		// +--------------------+-----------+------------+-----------+-------------+-----------------------------------------------------------------------------+
		// | 8623306-bin.008517 | 117732471 | Query      |  81413306 |   117732534 | BEGIN                                                                       |
		// | 8623306-bin.008517 | 117732534 | Table_map  |  81413306 |   117732590 | table_id: 2925 (zabbix.history)                                             |
		// | 8623306-bin.008517 | 117732590 | Write_rows |  81413306 |   117735750 | table_id: 2925 flags: STMT_END_F                                            |
		// | 8623306-bin.008517 | 117735750 | Table_map  |  81413306 |   117735810 | table_id: 2968 (zabbix.history_uint)                                        |
		// | 8623306-bin.008517 | 117735810 | Write_rows |  81413306 |   117744020 | table_id: 2968                                                              |
		// | 8623306-bin.008517 | 117744020 | Write_rows |  81413306 |   117752230 | table_id: 2968                                                              |
		// | 8623306-bin.008517 | 117752230 | Write_rows |  81413306 |   117755340 | table_id: 2968 flags: STMT_END_F                                            |
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
		//
		// Pos:117735810 will not be checkpointed
		// Pos:117752230 will be checkpointed
		//
		// otherwise when restart and resume from last checkpoint:
		// invalid table id 2413, no correspond table map event
		return nil
	}

	return m.commitPosition(r.Log, r.Position)
}

func (m *MySlave) commitPosition(file string, offset uint32) error {
	m.state.File = file
	m.state.Offset = offset
	return m.p.Commit(m.state)
}

// Events returns the iterator of mysql binlog rows event.
func (m *MySlave) Events() <-chan *model.RowsEvent {
	return m.rowsEvent
}

// Errors returns the iterator of unexpected errors.
func (m *MySlave) Errors() <-chan error {
	return m.errors
}
