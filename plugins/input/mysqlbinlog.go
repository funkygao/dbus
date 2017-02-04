package input

import (
	"encoding/json"

	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/plugins/input/mysqlbinlog"
	conf "github.com/funkygao/jsconf"
	"github.com/siddontang/go-mysql/canal"
)

var (
	_ canal.RowsEventHandler = &MysqlbinlogInput{}
)

type MysqlbinlogInput struct {
	stopChan chan struct{}
	binlog   chan []byte

	binlogStream *mysqlbinlog.MysqlBinlog
}

func (this *MysqlbinlogInput) Init(config *conf.Conf) {
	this.stopChan = make(chan struct{})
	this.binlog = make(chan []byte)
	this.binlogStream = mysqlbinlog.New().LoadConfig(config)
}

func (this *MysqlbinlogInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
	if err := this.binlogStream.Start(); err != nil {
		panic(err)
	}

	this.binlogStream.RegRowsEventHandler(this)

	for {
		select {
		case <-this.stopChan:
			return nil

		case pack, ok := <-r.InChan():
			if !ok {
				break
			}

			pack.Payload = <-this.binlog
			r.Inject(pack)
		}
	}

	return nil
}

func (this *MysqlbinlogInput) Stop() {
	close(this.stopChan)
}

func (this *MysqlbinlogInput) String() string {
	return "MysqlbinlogInput"
}

func (this *MysqlbinlogInput) Do(e *canal.RowsEvent) error {
	b, _ := json.Marshal(e)
	this.binlog <- b
	return nil
}

func init() {
	engine.RegisterPlugin("MysqlbinlogInput", func() engine.Plugin {
		return new(MysqlbinlogInput)
	})
}
