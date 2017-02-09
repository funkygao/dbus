package myslave

import (
	"fmt"
	"time"

	conf "github.com/funkygao/jsconf"
	log "github.com/funkygao/log4go"
	"github.com/siddontang/go-mysql/replication"
)

// MySlave is a mimic mysql slave that replicates binlog from
// mysql master using IO thread.
type MySlave struct {
	c *conf.Conf
	r *replication.BinlogSyncer
	p positioner
	m *slaveMetrics

	masterAddr string
	host       string
	port       uint16
	GTID       bool // global tx id

	dbExcluded, tableExcluded map[string]struct{}

	errors    chan error
	rowsEvent chan *RowsEvent
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

	m.p = newPositionerZk(m.c.String("zone", ""), m.masterAddr, m.c.Duration("pos_commit_interval", time.Second))
	m.m = newMetrics(fmt.Sprintf("dbus.myslave.%s", m.masterAddr))

	return m
}

func (m *MySlave) Close() {
	if m.r != nil {
		m.r.Close()
	}
	if err := m.p.Flush(); err != nil {
		log.Error("flush: %s", err)
	}
	//close(m.errors)
}

func (m *MySlave) StopReplication() {
	m.r.Close()
	if err := m.p.Flush(); err != nil {
		log.Error("flush: %s", err)
	}
}

func (m *MySlave) MarkAsProcessed(r *RowsEvent) error {
	return m.p.MarkAsProcessed(r.Log, r.Position)
}

func (m *MySlave) EventStream() <-chan *RowsEvent {
	return m.rowsEvent
}

func (m *MySlave) Errors() <-chan error {
	return m.errors
}
