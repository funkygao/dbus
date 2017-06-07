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

type myslaveLag struct {
	ident string

	zkzone  *zk.ZkZone
	stopper <-chan struct{}
	wg      *sync.WaitGroup

	addr, db string
}

func (this *myslaveLag) Init(ctx monitor.Context) {
	this.zkzone = ctx.ZkZone()
	this.stopper = ctx.StopChan()
	this.wg = ctx.Inflight()

	this.addr = ctx.InfluxAddr()
	this.db = ctx.InfluxDB()
}

func (this *myslaveLag) Run() {
	defer this.wg.Done()

	binlogLag := metrics.NewRegisteredGauge("_dbus.myslavelag", nil)

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-this.stopper:
			log.Info("%s stopped", this.ident)
			return

		case <-ticker.C:
			n, err := this.binlogLag(30)
			if err != nil {
				log.Error("%s: %s", this.ident, err)
			} else {
				binlogLag.Update(int64(n))
			}
		}
	}
}

func (this *myslaveLag) binlogLag(sec int) (int, error) {
	res, err := queryInfluxDB(this.addr, this.db,
		fmt.Sprintf(`SELECT * FROM "mysql.binlog.lag.gauge" WHERE time > now() - 1m AND value > %d`, sec))
	if err != nil {
		return 0, err
	}

	total := 0
	for _, row := range res {
		for _, x := range row.Series {
			total += len(x.Values)

			for _, val := range x.Values {
				// val: [time, input name, host, value]
				pluginName, _ := val[1].(string)
				host, _ := val[2].(string)
				log.Warn("[%s] on %s lag>%d", pluginName, host, sec)
			}
		}
	}

	return total, nil
}

func init() {
	monitor.RegisterWatcher("dbus.mylag", func() monitor.Watcher {
		return &myslaveLag{ident: "dbus.mylag"}
	})
}
