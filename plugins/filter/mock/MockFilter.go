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
	for pack := range r.InChan() {
		pack.Recycle()
	}

	return nil
}
