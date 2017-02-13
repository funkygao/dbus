package input

import (
	"fmt"
	"time"

	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/plugins/input/myslave"
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

	// so that KafkaOutput can reuse
	key := fmt.Sprintf("myslave.%s", config.String("name", ""))
	engine.Globals().Register(key, this.slave)
}

func (this *MysqlbinlogInput) Stop() {
	log.Trace("stopping...")
	close(this.stopChan)
	this.slave.StopReplication()
}

func (this *MysqlbinlogInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
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
				log.Trace("yes sir!")
				return nil

			case err := <-errors:
				log.Error("backoff %s: %v", backoff, err)
				this.slave.StopReplication()

				select {
				case <-time.After(backoff):
				case <-this.stopChan:
					log.Trace("yes sir!")
					return nil
				}
				goto RESTART_REPLICATION

			case pack, ok := <-r.InChan():
				if !ok {
					log.Trace("yes sir!")
					return nil
				}

				select {
				case err := <-errors:
					log.Error("backoff %s: %v", backoff, err)
					this.slave.StopReplication()

					select {
					case <-time.After(backoff):
					case <-this.stopChan:
						log.Trace("yes sir!")
						return nil
					}
					goto RESTART_REPLICATION

				case row, ok := <-rows:
					if !ok {
						log.Info("event stream closed")
						return nil
					}

					pack.Payload = row
					r.Inject(pack)

				case <-this.stopChan:
					log.Trace("yes sir!")
					return nil
				}
			}
		}
	}

	return nil
}

func init() {
	engine.RegisterPlugin("MysqlbinlogInput", func() engine.Plugin {
		return new(MysqlbinlogInput)
	})
}
