package redis

import (
	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/pkg/model"
	conf "github.com/funkygao/jsconf"
	log "github.com/funkygao/log4go"
)

type RedisbinlogInput struct {
	stopChan chan struct{}
}

func (this *RedisbinlogInput) Init(config *conf.Conf) {
	panic("Not implemented")

	this.stopChan = make(chan struct{})
}

func (this *RedisbinlogInput) OnAck(pack *engine.Packet) error {
	return nil
}

func (this *RedisbinlogInput) Stop(r engine.InputRunner) {
	log.Debug("[%s] stopping...", r.Name())
	close(this.stopChan)
}

func (this *RedisbinlogInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
	for {
		select {
		case <-this.stopChan:
			return nil

		case pack, ok := <-r.InChan():
			if !ok {
				log.Debug("yes sir!")
				break
			}

			pack.Payload = model.Bytes("hello world")
			r.Inject(pack)
		}
	}

	return nil
}
