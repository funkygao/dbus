package input

import (
	"fmt"
	"time"

	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/pkg/myslave"
	conf "github.com/funkygao/jsconf"
	log "github.com/funkygao/log4go"
)

var (
	_ engine.Input = &MysqlbinlogInput{}
)

// MysqlbinlogInput is an input plugin that pretends to be a mysql instance's
// slave and consumes mysql binlog events.
type MysqlbinlogInput struct {
	maxEventLength int
	stopChan       chan struct{}

	slave *myslave.MySlave
}

func (this *MysqlbinlogInput) Init(config *conf.Conf) {
	this.maxEventLength = config.Int("max_event_length", (1<<20)-100)
	this.stopChan = make(chan struct{})
	this.slave = myslave.New().LoadConfig(config)

	// so that KafkaOutput can reuse
	engine.Globals().Register(fmt.Sprintf("myslave.%s", config.String("name", "")), this.slave)
}

func (this *MysqlbinlogInput) Stop(r engine.InputRunner) {
	log.Trace("[%s] stopping...", r.Name())

	close(this.stopChan)
	this.slave.StopReplication()
}

func (this *MysqlbinlogInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
	name := r.Name()
	backoff := time.Second * 5

	for {
	RESTART_REPLICATION:

		log.Info("[%s] starting replication...", name)

		ready := make(chan struct{})
		go this.slave.StartReplication(ready)
		select {
		case <-ready:
		case <-this.stopChan:
			log.Trace("[%s] yes sir!", name)
			return nil
		}

		rows := this.slave.Events()
		errors := this.slave.Errors()
		for {
			select {
			case <-this.stopChan:
				log.Trace("[%s] yes sir!", name)
				return nil

			case err := <-errors:
				// e,g.
				// ERROR 1236 (HY000): Could not find first log file name in binary log index file
				// ERROR 1236 (HY000): Could not open log file
				log.Error("[%s] backoff %s: %v", name, backoff, err)
				this.slave.StopReplication()

				select {
				case <-time.After(backoff):
				case <-this.stopChan:
					log.Trace("[%s] yes sir!", name)
					return nil
				}
				goto RESTART_REPLICATION

			case pack, ok := <-r.InChan():
				if !ok {
					log.Trace("[%s] yes sir!", name)
					return nil
				}

				select {
				case err := <-errors:
					// TODO is this neccessary?
					log.Error("[%s] backoff %s: %v", name, backoff, err)
					this.slave.StopReplication()

					select {
					case <-time.After(backoff):
					case <-this.stopChan:
						log.Trace("[%s] yes sir!", name)
						return nil
					}
					goto RESTART_REPLICATION

				case row, ok := <-rows:
					if !ok {
						log.Info("[%s] event stream closed", name)
						return nil
					}

					if row.Length() < this.maxEventLength {
						pack.Payload = row
						r.Inject(pack)
					} else {
						log.Warn("[%s] ignored len=%d %s", name, row.Length(), row.MetaInfo())
						pack.Recycle()
					}

				case <-this.stopChan:
					log.Trace("[%s] yes sir!", name)
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
