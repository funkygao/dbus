package command

import (
	"github.com/funkygao/gafka/cmd/gk/command"
	"github.com/funkygao/gocli"
)

type Zones struct {
	Ui  cli.Ui
	Cmd string
}

func (this *Zones) Run(args []string) (exitCode int) {
	zones := command.Zones{
		Ui:  this.Ui,
		Cmd: this.Cmd,
	}
	return zones.Run(args)
}

func (*Zones) Synopsis() string {
	return "Print zones defined in $HOME/.gafka.cf"
}

func (this *Zones) Help() string {
	return ""
}
