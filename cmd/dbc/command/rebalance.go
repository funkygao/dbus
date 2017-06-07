package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gocli"
)

type Rebalance struct {
	Ui  cli.Ui
	Cmd string
}

func (this *Rebalance) Run(args []string) (exitCode int) {
	var (
		zone    string
		cluster string
	)
	cmdFlags := flag.NewFlagSet("rebalance", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&zone, "z", ctx.ZkDefaultZone(), "")
	cmdFlags.StringVar(&cluster, "c", "", "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	mgr := openClusterManager(zone)
	defer mgr.Close()

	if err := mgr.Rebalance(); err != nil {
		this.Ui.Error(err.Error())
	} else {
		this.Ui.Info("rebalanced")
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

    -c cluster

`, this.Cmd, this.Synopsis())
	return strings.TrimSpace(help)
}
