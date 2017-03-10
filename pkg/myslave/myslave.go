package myslave

import (
	"fmt"
	"time"

	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/pkg/model"
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

	name string

	Predicate func(schema, table string) bool
	GTID      bool // global tx id

	masterAddr   string
	host         string
	port         uint16
	user, passwd string

	dbAllowed  map[string]struct{}
	dbExcluded map[string]struct{}

	isMaster  bool
	errors    chan error
	rowsEvent chan *model.RowsEvent
}

// New creates a MySlave instance.
func New() *MySlave {
	return &MySlave{
		dbExcluded: map[string]struct{}{},
		dbAllowed:  map[string]struct{}{},
	}
}

// LoadConfig initialize internal state according to the config section.
func (m *MySlave) LoadConfig(config *conf.Conf) *MySlave {
	m.c = config

	var err error
	var dbs []string
	var zone string
	zone, m.host, m.port, m.user, m.passwd, dbs, err = ParseDSN(m.c.String("dsn", ""))
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
	m.p = newPositionerZk(m.name, m.z, m.masterAddr, m.c.Duration("pos_commit_interval", time.Second))

	return m
}

// MarkAsProcessed notifies the positioner that a certain binlog event
// has been successfully processed and should be committed.
func (m *MySlave) MarkAsProcessed(r *model.RowsEvent) error {
	return m.p.MarkAsProcessed(r.Log, r.Position)
}

// Events returns the iterator of mysql binlog rows event.
func (m *MySlave) Events() <-chan *model.RowsEvent {
	return m.rowsEvent
}

// Errors returns the iterator of unexpected errors.
func (m *MySlave) Errors() <-chan error {
	return m.errors
}
