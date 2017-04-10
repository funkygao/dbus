package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/funkygao/columnize"
	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/gocli"

	// bootstrap plugins
	_ "github.com/funkygao/dbus/plugins/filter"
	_ "github.com/funkygao/dbus/plugins/input"
	_ "github.com/funkygao/dbus/plugins/output"
)

type Plugins struct {
	Ui  cli.Ui
	Cmd string

	fn string
}

func (this *Plugins) Run(args []string) (exitCode int) {
	cmdFlags := flag.NewFlagSet("binlog", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&this.fn, "c", "", "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	e := engine.New(nil).LoadFrom(this.fn)
	lines := []string{"Plugin|Name|Class"}
	for _, ir := range e.InputRunners {
		lines = append(lines, fmt.Sprintf("Input|%s|%s", ir.Name(), ir.Class()))
	}
	for _, fr := range e.FilterRunners {
		lines = append(lines, fmt.Sprintf("Filter|%s|%s", fr.Name(), fr.Class()))
	}
	for _, or := range e.OutputRunners {
		lines = append(lines, fmt.Sprintf("Output|%s|%s", or.Name(), or.Class()))
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

    -c config location
      If empty, load from zookeeper
      zk location example:
      localhost:2181/dbus/conf

`, this.Cmd, this.Synopsis())
	return strings.TrimSpace(help)
}
