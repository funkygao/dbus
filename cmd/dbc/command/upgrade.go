package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gocli"
)

type Upgrade struct {
	Ui  cli.Ui
	Cmd string
}

func (this *Upgrade) Run(args []string) (exitCode int) {
	var zone string
	cmdFlags := flag.NewFlagSet("upgrade", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&zone, "z", ctx.ZkDefaultZone(), "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	mgr := openClusterManager(zone)
	defer mgr.Close()

	swallow(mgr.TriggerUpgrade())
	this.Ui.Info("ok")

	return
}

func (*Upgrade) Synopsis() string {
	return "Trigger hot upgrade of all dbusd binaries"
}

func (this *Upgrade) Help() string {
	help := fmt.Sprintf(`
Usage: %s upgrade

    %s
`, this.Cmd, this.Synopsis())
	return strings.TrimSpace(help)
}
