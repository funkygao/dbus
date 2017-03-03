package input

import (
	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/pkg/model"
	conf "github.com/funkygao/jsconf"
	log "github.com/funkygao/log4go"
)

var (
	_ engine.Input = &RedisbinlogInput{}
)

type RedisbinlogInput struct {
	stopChan chan struct{}
}

func (this *RedisbinlogInput) Init(config *conf.Conf) {
	panic("Not implemented")

	this.stopChan = make(chan struct{})
}

func (this *RedisbinlogInput) Stop() {
	log.Trace("stop called")
	close(this.stopChan)
}

func (this *RedisbinlogInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
	for {
		select {
		case <-this.stopChan:
			return nil

		case pack, ok := <-r.InChan():
			if !ok {
				log.Trace("yes sir!")
				break
			}

			pack.Payload = model.Bytes("hello world")
			r.Inject(pack)
		}
	}

	return nil
}

func init() {
	engine.RegisterPlugin("RedisbinlogInput", func() engine.Plugin {
		return new(RedisbinlogInput)
	})
}
