package plugins

import (
	"github.com/funkygao/dbus/engine"
	conf "github.com/funkygao/jsconf"
)

type MockOutput struct {
}

func (this *MockOutput) Init(config *conf.Conf) {}

func (this *MockOutput) Run(r engine.OutputRunner, h engine.PluginHelper) error {
	for {
		select {
		case pack, ok := <-r.InChan():
			if !ok {
				return nil
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
