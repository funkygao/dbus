package plugins

import (
	"github.com/funkygao/dbus/engine"
	conf "github.com/funkygao/jsconf"
)

type MockInput struct {
	stopChan chan struct{}
	ident    string
}

func (this *MockInput) Init(config *conf.Conf) {
	this.stopChan = make(chan struct{})
	this.ident = config.String("ident", "")
	if this.ident == "" {
		panic("empty ident")
	}
}

func (this *MockInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
	for {
		select {
		case <-this.stopChan:
			return nil

		case pack, ok := <-r.InChan():
			if !ok {
				break
			}

			// manipulate the pack
			r.Inject(pack)
		}
	}

	return nil
}

func (this *MockInput) Stop() {
	close(this.stopChan)
}

func init() {
	engine.RegisterPlugin("MockInput", func() engine.Plugin {
		return new(MockInput)
	})
}
