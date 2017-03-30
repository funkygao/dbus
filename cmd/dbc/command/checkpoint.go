package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gocli"
)

type Checkpoint struct {
	Ui  cli.Ui
	Cmd string

	zone string
}

func (this *Checkpoint) Run(args []string) (exitCode int) {
	cmdFlags := flag.NewFlagSet("checkpoint", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&this.zone, "z", ctx.ZkDefaultZone(), "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	return
}

func (*Checkpoint) Synopsis() string {
	return "Manages cluster checkpoint TODO"
}

func (this *Checkpoint) Help() string {
	help := fmt.Sprintf(`
Usage: %s checkpoint [options]

    %s

Options:

    -z zone

`, this.Cmd, this.Synopsis())
	return strings.TrimSpace(help)
}
