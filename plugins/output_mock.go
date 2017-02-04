package plugins

import (
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
	tick := time.NewTicker(time.Second * 10)
	defer tick.Stop()

	var n, lastN int64
	globals := engine.Globals()
	for {
		select {
		case pack, ok := <-r.InChan():
			if !ok {
				return nil
			}

			n++

			if !this.blackhole {
				globals.Printf("-> %s", pack)
			}

			pack.Recycle()

		case <-tick.C:
			globals.Printf("throughput %s/s", gofmt.Comma((n-lastN)/10))
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
