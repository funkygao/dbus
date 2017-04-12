package mysql

import (
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
	stopChan       chan struct{}

	slave *myslave.MySlave
	cf    *conf.Conf
}

func (this *MysqlbinlogInput) Init(config *conf.Conf) {
	this.maxEventLength = config.Int("max_event_length", (1<<20)-100)
	this.cf = config
	this.stopChan = make(chan struct{})
}

func (this *MysqlbinlogInput) Stop(r engine.InputRunner) {
	log.Debug("[%s] stopping...", r.Name())

	close(this.stopChan)
	if this.slave != nil {
		this.slave.StopReplication()
	}
}

func (this *MysqlbinlogInput) MySlave() *myslave.MySlave {
	return this.slave
}

// used only for dbc: ugly design
func (this *MysqlbinlogInput) ConnectMyslave(dsn string) {
	this.slave = myslave.New("", dsn, engine.Globals().ZrootCheckpoint).LoadConfig(this.cf)
}

func (this *MysqlbinlogInput) OnAck(pack *engine.Packet) error {
	return this.slave.MarkAsProcessed(pack.Payload.(*model.RowsEvent))
}

func (this *MysqlbinlogInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
	if dsn := r.Conf().String("dsn", ""); len(dsn) > 0 {
		return this.runStandalone(dsn, r, h)
	}

	return this.runClustered(r, h)
}
