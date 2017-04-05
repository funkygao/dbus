package watchers

import (
	"fmt"
	"sync"
	"time"

	"github.com/funkygao/gafka/cmd/kguard/monitor"
	"github.com/funkygao/gafka/zk"
	"github.com/funkygao/go-metrics"
	log "github.com/funkygao/log4go"
)

type dbusWatcher struct {
	ident string

	zkzone  *zk.ZkZone
	stopper <-chan struct{}
	wg      *sync.WaitGroup
}

func (this *dbusWatcher) Init(ctx monitor.Context) {
	this.zkzone = ctx.ZkZone()
	this.stopper = ctx.StopChan()
	this.wg = ctx.Inflight()
}

func (this *dbusWatcher) Run() {
	defer this.wg.Done()

	dbs := metrics.NewRegisteredGauge("dbus.myslave.db", nil)
	owner := metrics.NewRegisteredGauge("dbus.myslave.owenr", nil)
	runner := metrics.NewRegisteredGauge("dbus.myslave.runner", nil)
	orphanDbs := metrics.NewRegisteredGauge("dbus.myslave.orphan", nil)

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-this.stopper:
			log.Info("%s stopped", this.ident)
			return

		case <-ticker.C:
			dbN, ownerN, runnerN := this.fetchData()
			dbs.Update(int64(dbN))
			owner.Update(int64(ownerN))
			runner.Update(int64(runnerN))
			orphanDbs.Update(int64(dbN - ownerN))
		}
	}
}

func (this *dbusWatcher) fetchData() (dbN int, ownerN int, runnerN int) {
	root := "/dbus/myslave"
	dbs, _, err := this.zkzone.Conn().Children(root)
	if err != nil {
		log.Error("%s %s", root, err)
		return
	}
	dbN = len(dbs)

	for _, db := range dbs {
		idsPath := fmt.Sprintf("%s/%s/ids", root, db)
		runners, _, err := this.zkzone.Conn().Children(idsPath)
		if err != nil {
			log.Error("%s %s", idsPath, err)
			continue
		}
		runnerN += len(runners)

		ownerPath := fmt.Sprintf("%s/%s/owner", root, db)
		data, _, err := this.zkzone.Conn().Get(ownerPath)
		if err != nil {
			log.Error("%s %s", ownerPath, err)
		} else if len(data) == 0 {
			log.Error("empty master: %s", ownerPath)
		} else {
			ownerN += 1
		}
	}

	return
}

func init() {
	monitor.RegisterWatcher("dbus.dbus", func() monitor.Watcher {
		return &dbusWatcher{ident: "dbus.dbus"}
	})
}
