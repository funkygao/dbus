package mysql

import (
	"sync"

	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/pkg/model"
	"github.com/funkygao/dbus/pkg/myslave"
	conf "github.com/funkygao/jsconf"
)

// MysqlbinlogInput is an input plugin that pretends to be a mysql instance's
// slave and consumes mysql binlog events.
type MysqlbinlogInput struct {
	maxEventLength int
	cf             *conf.Conf
	clusterMode    bool

	mu     sync.RWMutex
	slaves []*myslave.MySlave
}

func (this *MysqlbinlogInput) Init(config *conf.Conf) {
	this.maxEventLength = config.Int("max_event_length", (1<<20)-100)
	this.cf = config
	if dsn := this.cf.String("dsn", ""); len(dsn) == 0 {
		this.clusterMode = true
	}
	this.slaves = make([]*myslave.MySlave, 0)
}

func (this *MysqlbinlogInput) OnAck(pack *engine.Packet) error {
	if !this.clusterMode {
		return this.slaves[0].MarkAsProcessed(pack.Payload.(*model.RowsEvent))
	}

	// cluster mode
	// FIXME
	// race condition: in cluster mode, when ACK, the slave might have been gone
	return this.slaves[0].MarkAsProcessed(pack.Payload.(*model.RowsEvent))
}

func (this *MysqlbinlogInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
	if this.clusterMode {
		return this.runClustered(r, h)
	}

	return this.runStandalone(this.cf.String("dsn", ""), r, h)
}
