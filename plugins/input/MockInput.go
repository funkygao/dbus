package input

import (
	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/pkg/model"
	conf "github.com/funkygao/jsconf"
	log "github.com/funkygao/log4go"
)

var (
	_ engine.Input = &MockInput{}
)

type MockInput struct {
	stopChan chan struct{}
}

func (this *MockInput) Init(config *conf.Conf) {
	this.stopChan = make(chan struct{})
}

func (this *MockInput) Stop(r engine.InputRunner) {
	log.Trace("[%s] stopping...", r.Name())
	close(this.stopChan)
}

func (this *MockInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
	payload := model.Bytes(`{"log":"6633343-bin.006419","pos":795931083,"db":"owl_t_prod_cd","tbl":"owl_mi","dml":"U","ts":1488934500,"rows":[[132332,"expired_keys","过期的key的个数","mock-monitor|member-mock|10.1.1.1|10489|expired_keys","mock-monitor",244526,"10489",null,"2015-12-25 17:12:00",null,null,"28571284","Ok","TypeLong","2017-03-08 08:54:01",null,null,null,null,null,null],[132332,"expired_keys","过期的key的个数","mock-monitor|member-mock|10.1.1.6|10489|expired_keys","mock-monitor",244526,"10489",null,"2015-12-25 17:12:00",null,null,"28571320","Ok","TypeLong","2017-03-08 08:55:00",null,null,null,null,null,null]]}`)
	for {
		select {
		case <-this.stopChan:
			return nil

		case pack, ok := <-r.InChan():
			if !ok {
				log.Trace("yes sir!")
				break
			}

			pack.Payload = payload
			r.Inject(pack)
		}
	}

	return nil
}

func init() {
	engine.RegisterPlugin("MockInput", func() engine.Plugin {
		return new(MockInput)
	})
}
