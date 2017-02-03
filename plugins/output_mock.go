package plugins

import (
	"log"

	"github.com/funkygao/dbus/engine"
	conf "github.com/funkygao/jsconf"
)

type MockOutput struct {
	blackhole bool
}

func (this *MockOutput) Init(config *conf.Conf) {
	this.blackhole = config.Bool("blackhole", false)
}

func (this *MockOutput) Run(r engine.OutputRunner, h engine.PluginHelper) error {
	for {
		select {
		case pack, ok := <-r.InChan():
			if !ok {
				return nil
			}

			if !this.blackhole {
				log.Printf("-> %s", string(pack.Payload))
			}

			pack.Recycle()
		}
	}

	return nil
}

func init() {
	engine.RegisterPlugin("MockOutput", func() engine.Plugin {
		return new(MockOutput)
	})
}
