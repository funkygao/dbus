package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/funkygao/gafka/ctx"
	hzk "github.com/funkygao/go-helix/store/zk"
	"github.com/funkygao/gocli"
)

type Clusters struct {
	Ui  cli.Ui
	Cmd string

	zone, cluster string
}

func (this *Clusters) Run(args []string) (exitCode int) {
	cmdFlags := flag.NewFlagSet("clusters", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&this.zone, "z", ctx.ZkDefaultZone(), "")
	cmdFlags.StringVar(&this.cluster, "add", "", "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	if len(this.cluster) != 0 {
		this.doAddCluster(ctx.ZoneZkAddrs(this.zone))
		return
	}

	return
}

func (this *Clusters) doAddCluster(zkSvr string) {
	admin := hzk.NewZkHelixAdmin(zkSvr)
	admin.Connect()

	if ok, err := admin.IsClusterSetup(this.cluster); err != nil {
		this.Ui.Error(err.Error())
		return
	} else if ok {
		this.Ui.Warnf("cluster %s already exists, skipped")
		return
	}

	admin.AddCluster(this.cluster)
	admin.AllowParticipantAutoJoin(this.cluster, true)
	this.Ui.Infof("cluster %s created", this.cluster)
}

func (*Clusters) Synopsis() string {
	return "Management of helix clusters"
}

func (this *Clusters) Help() string {
	help := fmt.Sprintf(`
Usage: %s clusters [options]

    %s

Options:

    -z zone

    -add cluster

`, this.Cmd, this.Synopsis())
	return strings.TrimSpace(help)
}
