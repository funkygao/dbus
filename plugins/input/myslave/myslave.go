package myslave

import (
	"fmt"
	"time"

	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/model"
	"github.com/funkygao/gafka/zk"
	conf "github.com/funkygao/jsconf"
	"github.com/siddontang/go-mysql/replication"
)

// MySlave is a mimic mysql slave that replicates binlog from mysql master using IO thread.
type MySlave struct {
	c *conf.Conf
	r *replication.BinlogSyncer
	p positioner
	m *slaveMetrics
	z *zk.ZkZone

	masterAddr string
	host       string
	port       uint16
	GTID       bool // global tx id

	dbExcluded, tableExcluded map[string]struct{}

	errors    chan error
	rowsEvent chan *model.RowsEvent
}

func New() *MySlave {
	return &MySlave{
		dbExcluded:    map[string]struct{}{},
		tableExcluded: map[string]struct{}{},
	}
}

func (m *MySlave) LoadConfig(config *conf.Conf) *MySlave {
	m.c = config

	m.host = m.c.String("master_host", "localhost")
	m.port = uint16(m.c.Int("master_port", 3306))
	m.masterAddr = fmt.Sprintf("%s:%d", m.host, m.port)
	m.GTID = m.c.Bool("GTID", false)
	for _, db := range config.StringList("db_excluded", nil) {
		m.dbExcluded[db] = struct{}{}
	}
	for _, table := range config.StringList("table_excluded", nil) {
		m.tableExcluded[table] = struct{}{}
	}

	m.m = newMetrics(m.host, m.port)
	zone := m.c.String("zone", "")
	if zone == "" {
		panic("zone required")
	}
	m.z = engine.Globals().GetOrRegisterZkzone(zone)
	m.p = newPositionerZk(m.z, m.masterAddr, m.c.Duration("pos_commit_interval", time.Second))

	return m
}

func (m *MySlave) MarkAsProcessed(r *model.RowsEvent) error {
	return m.p.MarkAsProcessed(r.Log, r.Position)
}

func (m *MySlave) EventStream() <-chan *model.RowsEvent {
	return m.rowsEvent
}

func (m *MySlave) Errors() <-chan error {
	return m.errors
}
