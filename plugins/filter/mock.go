package filter

import (
	"github.com/funkygao/dbus/engine"
	conf "github.com/funkygao/jsconf"
)

type MockFilter struct {
	ident string
}

func (this *MockFilter) Init(config *conf.Conf) {
	this.ident = config.String("ident", "")
	if this.ident == "" {
		panic("empty ident")
	}
}

func (this *MockFilter) Run(r engine.FilterRunner, h engine.PluginHelper) error {
	for {
		select {
		case pack, ok := <-r.InChan():
			if !ok {
				return nil
			}

			pack.Recycle()
		}
	}
}

func init() {
	engine.RegisterPlugin("MockFilter", func() engine.Plugin {
		return new(MockFilter)
	})
}
