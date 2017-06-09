package mysql

import (
	"sync"

	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/pkg/model"
	"github.com/funkygao/dbus/pkg/myslave"
	conf "github.com/funkygao/jsconf"
	log "github.com/funkygao/log4go"
)

// MysqlbinlogInput is an input plugin that pretends to be a mysql instance's
// slave and consumes mysql binlog events.
type MysqlbinlogInput struct {
	maxEventLength int
	cf             *conf.Conf
	clusterMode    bool

	mu     sync.RWMutex
	slaves map[string]*myslave.MySlave // cluster mode, key is DSN
	slave  *myslave.MySlave            // standalone mode
}

func (this *MysqlbinlogInput) Init(config *conf.Conf) {
	this.maxEventLength = config.Int("max_event_length", (1<<20)-100)
	this.cf = config
	if dsn := this.cf.String("dsn", ""); len(dsn) == 0 {
		this.clusterMode = true
		this.slaves = make(map[string]*myslave.MySlave)
	}
}

func (*MysqlbinlogInput) SampleConfig() string {
	return `
	max_event_length: 1048476
	recv_buffer: 524288
	server_id: 137
	semi_sync: false
	`
}

func (this *MysqlbinlogInput) Ack(pack *engine.Packet) error {
	if !this.clusterMode {
		return this.slave.MarkAsProcessed(pack.Payload.(*model.RowsEvent))
	}

	dsn := pack.Metadata.(string)
	this.mu.RLock()
	slave := this.slaves[dsn]
	this.mu.RUnlock()

	if slave == nil {
		log.Warn("unrecognized DSN: %s", dsn)
		return nil
	}

	// cluster mode
	// FIXME
	// race condition: in cluster mode, when ACK, the slave might have been gone
	return slave.MarkAsProcessed(pack.Payload.(*model.RowsEvent))
}

func (this *MysqlbinlogInput) End(r engine.InputRunner) {}

func (this *MysqlbinlogInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
	if this.clusterMode {
		return this.runClustered(r, h)
	}

	return this.runStandalone(this.cf.String("dsn", ""), r, h)
}
