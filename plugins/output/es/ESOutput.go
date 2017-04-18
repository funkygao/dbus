package es

import (
	"github.com/funkygao/dbus/engine"
	conf "github.com/funkygao/jsconf"
)

type ESOutput struct {
}

func (this *ESOutput) Init(config *conf.Conf) {
}

func (*ESOutput) SampleConfig() string {
	return ``
}

func (this *ESOutput) Run(r engine.OutputRunner, h engine.PluginHelper) error {
	for pack := range r.Exchange().InChan() {
		pack.Recycle()
	}

	return nil
}
