package plugins

import (
	"log"
	"time"

	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/golib/gofmt"
	conf "github.com/funkygao/jsconf"
)

type MockOutput struct {
	blackhole bool
}

func (this *MockOutput) Init(config *conf.Conf) {
	this.blackhole = config.Bool("blackhole", false)
}

func (this *MockOutput) Run(r engine.OutputRunner, h engine.PluginHelper) error {
	tick := time.NewTicker(time.Second * 5)
	defer tick.Stop()
	var n, lastN int64

	for {
		select {
		case pack, ok := <-r.InChan():
			if !ok {
				return nil
			}

			n++

			if !this.blackhole {
				log.Printf("-> %s", string(pack.Payload))
			}

			pack.Recycle()

		case <-tick.C:
			log.Printf("throughput %s/s", gofmt.Comma((n-lastN)/5))
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
