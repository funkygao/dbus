package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gocli"
)

type Participants struct {
	Ui  cli.Ui
	Cmd string

	zone string
}

func (this *Participants) Run(args []string) (exitCode int) {
	cmdFlags := flag.NewFlagSet("participants", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&this.zone, "z", ctx.ZkDefaultZone(), "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	mgr := openClusterManager(this.zone)
	defer mgr.Close()

	leader, err := mgr.Leader()
	if err != nil {
		this.Ui.Error(err.Error())
	}

	// list all resources
	ps, err := mgr.LiveParticipants()
	if err != nil {
		this.Ui.Error(err.Error())
		return
	}

	for _, p := range ps {
		if p.Equals(leader) {
			this.Ui.Warn(p.Endpoint)
		} else {
			this.Ui.Info(p.Endpoint)
		}
	}

	return
}

func (*Participants) Synopsis() string {
	return "Display live participants in cluster"
}

func (this *Participants) Help() string {
	help := fmt.Sprintf(`
Usage: %s participants [options]

    %s

Options:

    -z zone

`, this.Cmd, this.Synopsis())
	return strings.TrimSpace(help)
}
