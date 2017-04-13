package es

import (
	"github.com/funkygao/dbus/engine"
	conf "github.com/funkygao/jsconf"
)

type ESOutput struct {
}

func (this *ESOutput) Init(config *conf.Conf) {
}

func (this *ESOutput) Run(r engine.OutputRunner, h engine.PluginHelper) error {
	for {
		select {
		case <-h.Stopper():
			return nil

		case pack, ok := <-r.Exchange().InChan():
			if !ok {
				return nil
			}

			pack.Recycle()

		}
	}

	return nil
}
