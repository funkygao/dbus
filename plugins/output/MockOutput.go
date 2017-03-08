package output

import (
	"time"

	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/golib/gofmt"
	conf "github.com/funkygao/jsconf"
	log "github.com/funkygao/log4go"
)

var (
	_ engine.Output = &MockOutput{}
)

type MockOutput struct {
	blackhole bool
}

func (this *MockOutput) Init(config *conf.Conf) {
	this.blackhole = config.Bool("blackhole", false)
}

func (this *MockOutput) Run(r engine.OutputRunner, h engine.PluginHelper) error {
	tick := time.NewTicker(time.Second * 10)
	defer tick.Stop()

	var n, lastN int64
	name := r.Name()
	for {
		select {
		case pack, ok := <-r.InChan():
			if !ok {
				return nil
			}

			n++

			if !this.blackhole {
				log.Info("[%s] -> %s", name, pack)
			}

			pack.Recycle()

		case <-tick.C:
			log.Info("throughput %s/s", gofmt.Comma((n-lastN)/10))
			lastN = n
		}
	}

	return nil
}

func init() {
	engine.RegisterPlugin("MockOutput", func() engine.Plugin {
		return new(MockOutput)
	})
}
