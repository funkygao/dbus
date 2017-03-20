package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/plugins/input/mysql"
	"github.com/funkygao/gocli"

	// bootstrap plugins
	_ "github.com/funkygao/dbus/plugins/filter"
	_ "github.com/funkygao/dbus/plugins/input"
	_ "github.com/funkygao/dbus/plugins/output"
)

type Binlog struct {
	Ui  cli.Ui
	Cmd string

	fn          string
	plugin      string
	binlogsMode bool
	binlogMode  bool
}

func (this *Binlog) Run(args []string) (exitCode int) {
	cmdFlags := flag.NewFlagSet("binlog", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&this.fn, "c", "", "")
	cmdFlags.StringVar(&this.plugin, "id", "", "")
	cmdFlags.BoolVar(&this.binlogsMode, "logs", false, "")
	cmdFlags.BoolVar(&this.binlogMode, "pos", false, "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	e := engine.New(nil).LoadConfigFile(this.fn)

	switch {
	case this.binlogMode:
		return

	case this.binlogsMode:
		ir, present := e.InputRunners[this.plugin]
		if !present {
			this.Ui.Errorf("%s not found", this.plugin)
			return 2
		}

		my := ir.Plugin().(*mysql.MysqlbinlogInput)
		logs, err := my.MySlave().MasterBinlogs()
		if err != nil {
			panic(err)
		}

		this.Ui.Info(this.plugin)
		for _, l := range logs {
			this.Ui.Output(l)
		}
		return

	default:
		// list all MysqlBinlogInput plugin names
		this.Ui.Outputf("All MysqlbinlogInput plugins:")
		for name, ir := range e.InputRunners {
			if _, ok := ir.Plugin().(*mysql.MysqlbinlogInput); ok {
				this.Ui.Info(name)
			}
		}
	}

	return
}

func (*Binlog) Synopsis() string {
	return "Display mysql binlog related info"
}

func (this *Binlog) Help() string {
	help := fmt.Sprintf(`
Usage: %s binlog -c filename [options]

    %s

    Must run on the same host as dbusd.

Options:

    -id input plugin name

    -logs
      Display a master binlog status

    -pos file:offset
      Print a single mysql binlog event content

`, this.Cmd, this.Synopsis())
	return strings.TrimSpace(help)
}
