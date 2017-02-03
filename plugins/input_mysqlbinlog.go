package plugins

import (
	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/plugins/mysqlbinlog"
	conf "github.com/funkygao/jsconf"
)

type MysqlbinlogInput struct {
	stopChan chan struct{}
	ident    string

	binlogStream *mysqlbinlog.MysqlBinlog
}

func (this *MysqlbinlogInput) Init(config *conf.Conf) {
	this.stopChan = make(chan struct{})
	this.ident = config.String("ident", "")
	if this.ident == "" {
		panic("empty ident")
	}

	this.binlogStream = mysqlbinlog.New()
}

func (this *MysqlbinlogInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
	for {
		select {
		case <-this.stopChan:
			return nil

		case pack, ok := <-r.InChan():
			if !ok {
				break
			}

			r.Inject(pack)
		}
	}

	return nil
}

func (this *MysqlbinlogInput) Stop() {
	close(this.stopChan)
}

func init() {
	engine.RegisterPlugin("MysqlbinlogInput", func() engine.Plugin {
		return new(MysqlbinlogInput)
	})
}
