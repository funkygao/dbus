package redis

import (
	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/pkg/model"
	conf "github.com/funkygao/jsconf"
)

type RedisbinlogInput struct {
}

func (this *RedisbinlogInput) Init(config *conf.Conf) {
	panic("Not implemented")
}

func (this *RedisbinlogInput) OnAck(pack *engine.Packet) error {
	return nil
}

func (this *RedisbinlogInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
	ex := r.Exchange()
	stopper := r.Stopper()
	for {
		select {
		case <-stopper:
			return nil

		case pack := <-ex.InChan():
			pack.Payload = model.Bytes("hello world")
			ex.Inject(pack)
		}
	}

	return nil
}
