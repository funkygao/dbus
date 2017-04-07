package watchers

import (
	"sync"
	"time"

	"github.com/funkygao/dbus/pkg/cluster"
	czk "github.com/funkygao/dbus/pkg/cluster/zk"
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

	resourcesGauge := metrics.NewRegisteredGauge("dbus.resources", nil)
	orphanResourcesGauge := metrics.NewRegisteredGauge("dbus.resources.orphan", nil)
	participantsGauge := metrics.NewRegisteredGauge("dbus.participants", nil)
	reelectGauge := metrics.NewRegisteredGauge("dbus.reelect", nil)

	mgr := czk.NewManager(this.zkzone.ZkAddrs())
	if err := mgr.Open(); err != nil {
		log.Warn("%s quit: %v", this.ident, err)
		return
	}

	lastLeader := cluster.Participant{}
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-this.stopper:
			log.Info("%s stopped", this.ident)
			return

		case <-ticker.C:
			orphanN := 0
			if resources, err := mgr.RegisteredResources(); err != nil {
				log.Error("%s %v", this.ident, err)
			} else {
				for _, r := range resources {
					if r.IsOrphan() {
						orphanN++
					}
				}

				resourcesGauge.Update(int64(len(resources)))
				orphanResourcesGauge.Update(int64(orphanN))
			}

			if liveParticipants, err := mgr.LiveParticipants(); err != nil {
				log.Error("%s %v", this.ident, err)
			} else {
				participantsGauge.Update(int64(len(liveParticipants)))
			}

			if leader, err := mgr.Leader(); err != nil {
				log.Error("%s %v", this.ident, err)
			} else {
				if lastLeader.Valid() && !leader.Equals(lastLeader) {
					log.Trace("%s new leader: %s", this.ident, leader)
					reelectGauge.Update(1)
				} else {
					reelectGauge.Update(0)
				}

				lastLeader = leader
			}
		}
	}
}

func init() {
	monitor.RegisterWatcher("dbus.dbus", func() monitor.Watcher {
		return &dbusWatcher{ident: "dbus.dbus"}
	})
}
