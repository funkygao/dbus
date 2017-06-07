package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/funkygao/columnize"
	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gafka/zk"
	"github.com/funkygao/gocli"

	// bootstrap plugins
	_ "github.com/funkygao/dbus/plugins/filter"
	_ "github.com/funkygao/dbus/plugins/input"
	_ "github.com/funkygao/dbus/plugins/output"
)

type Plugins struct {
	Ui  cli.Ui
	Cmd string

	fn      string
	longFmt bool
}

func (this *Plugins) Run(args []string) (exitCode int) {
	var zone, cluster string
	cmdFlags := flag.NewFlagSet("binlog", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&this.fn, "cf", "", "")
	cmdFlags.StringVar(&zone, "z", ctx.DefaultZone(), "")
	cmdFlags.StringVar(&cluster, "c", "", "")
	cmdFlags.BoolVar(&this.longFmt, "l", false, "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	zkzone := zk.NewZkZone(zk.DefaultConfig(zone, ctx.ZoneZkAddrs(zone)))
	if len(cluster) == 0 {
		if cluster = zkzone.DefaultDbusCluster(); cluster == "" {
			this.Ui.Error("-c required")
			return
		}
	}

	globals := engine.DefaultGlobals()
	globals.Zone = zone
	globals.Cluster = cluster

	e := engine.New(globals).LoadFrom(this.fn)
	if this.longFmt {
		for _, ir := range e.InputRunners {
			this.Ui.Infof("%s(%s)", ir.Name(), ir.Class())
			for _, item := range ir.SampleConfigItems() {
				this.Ui.Outputf("    %s", item)
			}
		}
		for _, fr := range e.FilterRunners {
			this.Ui.Infof("%s(%s)", fr.Name(), fr.Class())
			for _, item := range fr.SampleConfigItems() {
				this.Ui.Outputf("    %s", item)
			}
		}
		for _, or := range e.OutputRunners {
			this.Ui.Infof("%s(%s)", or.Name(), or.Class())
			for _, item := range or.SampleConfigItems() {
				this.Ui.Outputf("    %s", item)
			}
		}

		return
	}

	lines := []string{"Plugin|Name|Class|Configuration"}
	for _, ir := range e.InputRunners {
		lines = append(lines, fmt.Sprintf("Input|%s|%s|%v", ir.Name(), ir.Class(), ir.Conf().Content()))
	}
	for _, fr := range e.FilterRunners {
		lines = append(lines, fmt.Sprintf("Filter|%s|%s|%v", fr.Name(), fr.Class(), fr.Conf().Content()))
	}
	for _, or := range e.OutputRunners {
		lines = append(lines, fmt.Sprintf("Output|%s|%s|%v", or.Name(), or.Class(), or.Conf().Content()))
	}

	if len(lines) > 1 {
		this.Ui.Output(columnize.SimpleFormat(lines))
	}

	return
}

func (*Plugins) Synopsis() string {
	return "List all plugins"
}

func (this *Plugins) Help() string {
	help := fmt.Sprintf(`
Usage: %s plugins [options]

    %s

Options:

    -z zone

    -c cluster

    -cf config location
      If empty, load from zookeeper
      zk location example:
      localhost:2181/dbus/conf

    -l
      Use a long listing format.

`, this.Cmd, this.Synopsis())
	return strings.TrimSpace(help)
}
