package myslave

import (
	"fmt"

	conf "github.com/funkygao/jsconf"
	"github.com/siddontang/go-mysql/replication"
)

// MySlave is a mimic mysql slave that replicates binlog from
// mysql master using IO thread.
type MySlave struct {
	r *replication.BinlogSyncer

	gtid       bool // TODO
	masterAddr string
	pos        *checkpoint

	errors    chan error
	rowsEvent chan *RowsEvent
}

func New() *MySlave {
	return &MySlave{
		pos: &checkpoint{name: "master.info"},
	}
}

func (m *MySlave) LoadConfig(config *conf.Conf) *MySlave {
	m.gtid = config.Bool("gtid", false)
	host := config.String("master_host", "localhost")
	port := uint16(config.Int("master_port", 3306))
	m.masterAddr = fmt.Sprintf("%s:%d", host, port)
	m.r = replication.NewBinlogSyncer(&replication.BinlogSyncerConfig{
		ServerID: uint32(config.Int("server_id", 1007)),
		Flavor:   config.String("flavor", "mysql"), // or mariadb
		Host:     host,
		Port:     port,
		User:     config.String("user", "root"),
		Password: config.String("password", ""),
	})

	var err error
	if m.pos, err = loadCheckpoint("master.info"); err != nil {
		panic(err)
	}
	if len(m.pos.Addr) != 0 && m.masterAddr != m.pos.Addr {
		// master changed, reset
		m.pos = &checkpoint{Addr: m.masterAddr}
	}

	return m
}

func (m *MySlave) Close() {
	m.pos.Save(true)
	m.r.Close()
	close(m.errors)
}

func (m *MySlave) MarkAsProcessed(r *RowsEvent) error {
	m.pos.Update(r.Name, r.Position)
	return m.pos.Save(false)
}

func (m *MySlave) Checkpoint() error {
	return m.pos.Save(true)
}

func (m *MySlave) EventStream() <-chan *RowsEvent {
	return m.rowsEvent
}

func (m *MySlave) Errors() <-chan error {
	return m.errors
}
