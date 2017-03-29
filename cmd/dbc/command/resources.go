package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/funkygao/columnize"
	"github.com/funkygao/dbus/pkg/cluster"
	czk "github.com/funkygao/dbus/pkg/cluster/zk"
	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gocli"
)

type Resources struct {
	Ui  cli.Ui
	Cmd string

	zone        string
	addResource string
}

func (this *Resources) Run(args []string) (exitCode int) {
	cmdFlags := flag.NewFlagSet("resources", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&this.zone, "z", ctx.ZkDefaultZone(), "")
	cmdFlags.StringVar(&this.addResource, "add", "", "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	controller := czk.NewStandalone(ctx.ZoneZkAddrs(this.zone))

	if len(this.addResource) != 0 {
		tuples := strings.SplitN(this.addResource, "-", 2)
		if len(tuples) != 2 {
			this.Ui.Error("invalid resource fmt")
			return 2
		}

		this.doAddResource(controller, tuples[0], tuples[1])
		return
	}

	// list all resources
	m, err := controller.RegisteredResources()
	if err != nil {
		this.Ui.Error(err.Error())
		return
	}

	lines := []string{"InputPlugin|Resources"}
	for input, resources := range m {
		lines = append(lines, fmt.Sprintf("%s|%+v", input, resources))
	}
	if len(lines) > 1 {
		this.Ui.Output(columnize.SimpleFormat(lines))
	}

	return
}

func (this *Resources) doAddResource(controller cluster.Controller, input, resource string) {
	if err := controller.RegisterResource(input, resource); err != nil {
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

`, this.Cmd, this.Synopsis())
	return strings.TrimSpace(help)
}
