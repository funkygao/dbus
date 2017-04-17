package mysql

import (
	"time"

	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/pkg/myslave"
	log "github.com/funkygao/log4go"
)

func (this *MysqlbinlogInput) runStandalone(dsn string, r engine.InputRunner, h engine.PluginHelper) error {
	defer func() {
		this.mu.RLock()
		slave := this.slaves[0]
		this.mu.RUnlock()

		if slave != nil {
			slave.StopReplication()
		}
	}()

	globals := engine.Globals()
	ex := r.Exchange()
	name := r.Name()
	backoff := time.Second * 5
	stopper := r.Stopper()

	this.mu.Lock()
	this.slaves = this.slaves[:0]
	this.slaves = append(this.slaves, myslave.New(name, dsn, globals.ZrootCheckpoint).LoadConfig(this.cf))
	theSlave := this.slaves[0]
	this.mu.Unlock()

	if err := theSlave.AssertValidRowFormat(); err != nil {
		panic(err)
	}

	for {
	RESTART_REPLICATION:

		log.Trace("[%s] starting replication from %s...", name, dsn)

		if img, err := theSlave.BinlogRowImage(); err != nil {
			log.Error("[%s] %v", name, err)
		} else {
			log.Trace("[%s] binlog row image=%s", name, img)
		}

		ready := make(chan struct{})
		go theSlave.StartReplication(ready)
		select {
		case <-ready:
		case <-stopper:
			log.Debug("[%s] yes sir!", name)
			return nil
		}

		rows := theSlave.Events()
		errors := theSlave.Errors()
		for {
			select {
			case <-stopper:
				log.Debug("[%s] yes sir!", name)
				return nil

			case err := <-errors:
				// e,g.
				// ERROR 1236 (HY000): Could not find first log file name in binary log index file
				// ERROR 1236 (HY000): Could not open log file
				log.Error("[%s] backoff %s: %v, stop from %s", name, backoff, err, dsn)
				theSlave.StopReplication()

				select {
				case <-time.After(backoff):
				case <-stopper:
					log.Debug("[%s] yes sir!", name)
					return nil
				}
				goto RESTART_REPLICATION

			case pack := <-ex.InChan():
				select {
				case err := <-errors:
					log.Error("[%s] backoff %s: %v, stop from %s", name, backoff, err, dsn)
					theSlave.StopReplication()

					select {
					case <-time.After(backoff):
					case <-stopper:
						log.Debug("[%s] yes sir!", name)
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
						ex.Inject(pack)
					} else {
						// TODO this.slave.MarkAsProcessed(r), also consider batcher partial failure
						log.Warn("[%s] ignored len=%d %s", name, row.Length(), row.MetaInfo())
						pack.Recycle()
					}

				case <-stopper:
					log.Debug("[%s] yes sir!", name)
					return nil
				}
			}
		}
	}

	return nil
}
