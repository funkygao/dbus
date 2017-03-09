package mock

import (
	"github.com/funkygao/dbus/engine"
	conf "github.com/funkygao/jsconf"
)

type MockFilter struct {
}

func (this *MockFilter) Init(config *conf.Conf) {
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
