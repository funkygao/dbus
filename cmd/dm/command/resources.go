package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/go-helix"
	hzk "github.com/funkygao/go-helix/store/zk"
	"github.com/funkygao/gocli"
)

type Resources struct {
	Ui  cli.Ui
	Cmd string

	zone, cluster string
	addResource   string
}

func (this *Resources) Run(args []string) (exitCode int) {
	cmdFlags := flag.NewFlagSet("resources", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&this.zone, "z", ctx.ZkDefaultZone(), "")
	cmdFlags.StringVar(&this.cluster, "c", "", "")
	cmdFlags.StringVar(&this.addResource, "add", "", "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	if len(this.addResource) != 0 {
		if len(this.cluster) == 0 {
			this.Ui.Error("-c cluster is required")
			this.Ui.Output(this.Help())
			return
		}

		this.doAddResource(ctx.ZoneZkAddrs(this.zone))
		return
	}

	return
}

func (this *Resources) doAddResource(zkSvr string) {
	admin := hzk.NewZkHelixAdmin(zkSvr)
	admin.Connect()
	if ok, err := admin.IsClusterSetup(this.cluster); !ok || err != nil {
		this.Ui.Errorf("cluster %s not setup", this.cluster)
		return
	}

	partitions := 1
	resourceOption := helix.DefaultAddResourceOption(partitions, helix.StateModelOnlineOffline)
	resourceOption.RebalancerMode = helix.RebalancerModeFullAuto
	admin.AddResource(this.cluster, this.addResource, resourceOption)
	admin.Rebalance(this.cluster, this.addResource, 1)
	this.Ui.Infof("%s for cluster %s added and rebalanced", this.addResource, this.cluster)
}

func (*Resources) Synopsis() string {
	return "Manipulate helix resources in a cluster"
}

func (this *Resources) Help() string {
	help := fmt.Sprintf(`
Usage: %s resources [options]

    %s

Options:

    -z zone

    -c cluster

    -add resource

`, this.Cmd, this.Synopsis())
	return strings.TrimSpace(help)
}
