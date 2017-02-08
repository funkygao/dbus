package input

import (
	"time"

	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/plugins/input/myslave"
	//"github.com/funkygao/dbus/plugins/model"
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
}

func (this *MysqlbinlogInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
	defer this.slave.Close()

	backoff := time.Second * 5
	for {
	RESTART_REPLICATION:

		log.Info("starting replication")

		ready := make(chan struct{})
		go this.slave.StartReplication(ready)
		<-ready

		rows := this.slave.EventStream()
		errors := this.slave.Errors()
		for {
			select {
			case <-this.stopChan:
				log.Trace("yes sir! I quit")
				return nil

			case err := <-errors:
				log.Error("backoff %s: %v", backoff, err)
				this.slave.StopReplication()
				time.Sleep(backoff)
				goto RESTART_REPLICATION

			case pack, ok := <-r.InChan():
				if !ok {
					return nil
				}

				select {
				case err := <-errors:
					log.Error("backoff %s: %v", backoff, err)
					this.slave.StopReplication()
					time.Sleep(time.Second * 5)
					goto RESTART_REPLICATION

				case row, ok := <-rows:
					if !ok {
						log.Info("event stream closed")
						return nil
					}

					pack.Payload = row
					r.Inject(pack)

				case <-this.stopChan:
					log.Trace("yes sir! I quit")
					return nil
				}
			}
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
