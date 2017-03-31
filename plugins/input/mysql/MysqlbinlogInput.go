package mysql

import (
	"time"

	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/pkg/cluster"
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
	log.Trace("[%s] stopping...", r.Name())

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
	this.slave = myslave.New(dsn).LoadConfig(this.cf)
}

func (this *MysqlbinlogInput) OnAck(pack *engine.Packet) error {
	return this.slave.MarkAsProcessed(pack.Payload.(*model.RowsEvent))
}

func (this *MysqlbinlogInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
	name := r.Name()
	backoff := time.Second * 5

	var myResources []cluster.Resource
	resourcesCh := r.Resources()

	for {
	RESTART_REPLICATION:

		// wait till got some resource
		for {
			if len(myResources) != 0 {
				log.Trace("[%s] bingo! %d: %+v", name, len(myResources), myResources)
				break
			}

			log.Trace("[%s] awaiting resources", name)

			select {
			case <-this.stopChan:
				log.Trace("[%s] yes sir!", name)
				return nil
			case myResources = <-resourcesCh:
			}
		}

		dsn := myResources[0].DSN() // MysqlbinlogInput only consumes 1 resource
		log.Trace("[%s] starting replication from %+v...", name, dsn)
		this.slave = myslave.New(dsn).LoadConfig(this.cf)
		if err := this.slave.AssertValidRowFormat(); err != nil {
			panic(err)
		}

		if img, err := this.slave.BinlogRowImage(); err != nil {
			log.Error("[%s] %v", name, err)
		} else {
			log.Trace("[%s] binlog row image=%s", name, img)
		}

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
				// read initial handshake error, caused by Too many connections
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
					// TODO is this necessary?
					log.Error("[%s] backoff %s: %v", name, backoff, err)
					this.slave.StopReplication()

					select {
					case <-time.After(backoff):
					case <-this.stopChan:
						log.Trace("[%s] yes sir!", name)
						return nil
					}
					goto RESTART_REPLICATION

				case myResources = <-resourcesCh:
					log.Trace("[%s] cluster rebalanced", name)
					this.slave.StopReplication()
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
						// TODO this.slave.MarkAsProcessed(r), also consider batcher partial failure
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
