package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/funkygao/columnize"
	"github.com/funkygao/dbus/pkg/cluster"
	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gocli"
)

type Resources struct {
	Ui  cli.Ui
	Cmd string

	zone        string
	addResource string
	delResource string
}

func (this *Resources) Run(args []string) (exitCode int) {
	cmdFlags := flag.NewFlagSet("resources", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&this.zone, "z", ctx.ZkDefaultZone(), "")
	cmdFlags.StringVar(&this.addResource, "add", "", "")
	cmdFlags.StringVar(&this.delResource, "del", "", "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	mgr := openClusterManager(this.zone)
	defer mgr.Close()

	if len(this.addResource) != 0 {
		tuples := strings.SplitN(this.addResource, "-", 2)
		if len(tuples) != 2 {
			this.Ui.Error("invalid resource fmt")
			return 2
		}

		this.doAddResource(mgr, tuples[0], tuples[1])
		return
	}

	if len(this.delResource) > 0 {
		this.doDelResource(mgr, this.delResource)
		return
	}

	// list all resources
	resources, err := mgr.RegisteredResources()
	if err != nil {
		this.Ui.Error(err.Error())
		return
	}

	lines := []string{"InputPlugin|Resources|Epoch|Owner"}
	for _, res := range resources {
		if res.State.IsOrphan() {
			lines = append(lines, fmt.Sprintf("%s|%s|-|-", res.InputPlugin, res.Name))
		} else {
			lines = append(lines, fmt.Sprintf("%s|%s|%d|%s", res.InputPlugin, res.Name, res.State.LeaderEpoch, res.State.Owner))
		}
	}
	if len(lines) > 1 {
		this.Ui.Output(columnize.SimpleFormat(lines))
	}

	return
}

func (this *Resources) doDelResource(mgr cluster.Manager, resource string) {
	res := cluster.Resource{
		Name: resource,
	}
	if err := mgr.UnregisterResource(res); err != nil {
		this.Ui.Error(err.Error())
	} else {
		this.Ui.Info("ok")
	}
}

func (this *Resources) doAddResource(mgr cluster.Manager, input, resource string) {
	res := cluster.Resource{
		Name:        resource,
		InputPlugin: input,
	}
	if err := mgr.RegisterResource(res); err != nil {
		this.Ui.Error(err.Error())
	} else {
		this.Ui.Info("ok")
	}
}

func (*Resources) Synopsis() string {
	return "Define cluster resources"
}

func (this *Resources) Help() string {
	help := fmt.Sprintf(`
Usage: %s resources [options]

    %s

Options:

    -z zone

    -add input-resource
      resource DSN
      mysql zone://user:pass@host:port/db1,db2,...,dbn
      kafka zone://cluster/topic#partition

    -del resource

`, this.Cmd, this.Synopsis())
	return strings.TrimSpace(help)
}
