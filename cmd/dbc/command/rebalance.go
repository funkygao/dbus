package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/funkygao/gocli"
)

type Rebalance struct {
	Ui  cli.Ui
	Cmd string
}

func (this *Rebalance) Run(args []string) (exitCode int) {
	var zone string
	cmdFlags := flag.NewFlagSet("rebalance", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&zone, "z", "", "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	return
}

func (*Rebalance) Synopsis() string {
	return "Make cluster re-elect leader and rebalance"
}

func (this *Rebalance) Help() string {
	help := fmt.Sprintf(`
Usage: %s rebalance [options]

    %s

Options:

    -z zone

`, this.Cmd, this.Synopsis())
	return strings.TrimSpace(help)
}
