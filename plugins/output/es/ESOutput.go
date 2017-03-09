package output

import (
	"github.com/funkygao/dbus/engine"
	conf "github.com/funkygao/jsconf"
)

var (
	_ engine.Output = &ESOutput{}
)

type ESOutput struct {
}

func (this *ESOutput) Init(config *conf.Conf) {
}

func (this *ESOutput) Run(r engine.OutputRunner, h engine.PluginHelper) error {
	for {
		select {
		case pack, ok := <-r.InChan():
			if !ok {
				return nil
			}

			pack.Recycle()

		}
	}

	return nil
}

func init() {
	engine.RegisterPlugin("ESOutput", func() engine.Plugin {
		return new(ESOutput)
	})
}
