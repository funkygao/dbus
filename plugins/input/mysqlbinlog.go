package input

import (
	"time"

	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/plugins/input/myslave"
	"github.com/funkygao/dbus/plugins/model"
	conf "github.com/funkygao/jsconf"
	log "github.com/funkygao/log4go"
)

type MysqlbinlogInput struct {
	stopChan chan struct{}

	slave *myslave.MySlave
}

func (this *MysqlbinlogInput) Init(config *conf.Conf) {
	this.stopChan = make(chan struct{})
	this.slave = myslave.New().LoadConfig(config)
	logLevel := config.String("loglevel", "trace")
	for _, filter := range log.Global {
		filter.Level = log.ToLogLevel(logLevel, log.TRACE)
	}
}

func (this *MysqlbinlogInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
	for {
		ready := make(chan struct{})
		go this.slave.StartReplication(ready)
		<-ready

		rows := this.slave.EventStream()
		errors := this.slave.Errors()
		for {
			select {
			case <-this.stopChan:
				return nil

			case err := <-errors:
				// e,g. replication conn broken
				log.Error("slave: %s", err)
				time.Sleep(time.Second * 5)

			case pack, ok := <-r.InChan():
				if !ok {
					break
				}

				select {
				case row, ok := <-rows:
					if !ok {
						log.Info("event stream closed")
						return nil
					}

					pack.Payload = model.Bytes(row.Bytes())
					r.Inject(pack)

				case <-this.stopChan:
					return nil
				}
			}
		}
	}

	return nil
}

func (this *MysqlbinlogInput) Stop() {
	this.slave.Close()
	close(this.stopChan)
}

func init() {
	engine.RegisterPlugin("MysqlbinlogInput", func() engine.Plugin {
		return new(MysqlbinlogInput)
	})
}
