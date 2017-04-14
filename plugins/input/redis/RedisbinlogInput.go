package redis

import (
	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/pkg/model"
	conf "github.com/funkygao/jsconf"
	log "github.com/funkygao/log4go"
)

type RedisbinlogInput struct {
}

func (this *RedisbinlogInput) Init(config *conf.Conf) {
	panic("Not implemented")
}

func (this *RedisbinlogInput) OnAck(pack *engine.Packet) error {
	return nil
}

func (this *RedisbinlogInput) Stop(r engine.InputRunner) {
	log.Debug("[%s] stopping...", r.Name())
}

func (this *RedisbinlogInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
	ex := r.Exchange()
	stopper := h.Stopper()
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
