package myslave

import (
	"time"

	"github.com/funkygao/gafka/telemetry"
	"github.com/funkygao/gafka/telemetry/influxdb"
	"github.com/funkygao/go-metrics"
	conf "github.com/funkygao/jsconf"
	log "github.com/funkygao/log4go"
	"github.com/siddontang/go-mysql/replication"
)

// MySlave is a mimic mysql slave that replicates binlog from
// mysql master using IO thread.
type MySlave struct {
	r  *replication.BinlogSyncer
	cf *conf.Conf

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
	m.cf = config

	if cf, err := influxdb.NewConfig(m.cf.String("influx_addr", ""),
		m.cf.String("influx_db", "myslave"), "", "",
		m.cf.Duration("influx_tick", time.Minute)); err == nil {
		telemetry.Default = influxdb.New(metrics.DefaultRegistry, cf)
		go func() {
			if err := telemetry.Default.Start(); err != nil {
				log.Error("telemetry[%s]: %s", telemetry.Default.Name(), err)
			}
		}()
	} else {
		log.Warn("telemetry disabled for: %s", err)
	}

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
	//close(m.errors)
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
