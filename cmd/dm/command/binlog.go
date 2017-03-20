package command

import (
	"flag"
	"fmt"
	"strconv"
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
	binlogPos   string
}

func (this *Binlog) Run(args []string) (exitCode int) {
	cmdFlags := flag.NewFlagSet("binlog", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&this.fn, "c", "", "")
	cmdFlags.StringVar(&this.plugin, "id", "", "")
	cmdFlags.BoolVar(&this.binlogsMode, "logs", false, "")
	cmdFlags.StringVar(&this.binlogPos, "pos", "", "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	e := engine.New(nil).LoadConfigFile(this.fn)

	switch {
	case this.binlogPos != "":
		ir, present := e.InputRunners[this.plugin]
		if !present {
			this.Ui.Errorf("%s not found", this.plugin)
			return 2
		}

		parts := strings.SplitN(this.binlogPos, ":", 2)
		if len(parts) != 2 {
			this.Ui.Output(this.Help())
			return 2
		}
		pos, _ := strconv.Atoi(parts[1])

		my := ir.Plugin().(*mysql.MysqlbinlogInput)
		res, err := my.MySlave().BinlogByPos(parts[0], pos)
		if err != nil {
			panic(err)
		}

		// +--------------+-----+-------------+-----------+-------------+---------------------------------------+
		// | Log_name     | Pos | Event_type  | Server_id | End_log_pos | Info                                  |
		// +--------------+-----+-------------+-----------+-------------+---------------------------------------+
		// | mysql.000005 |   4 | Format_desc |         1 |         120 | Server ver: 5.6.23-log, Binlog ver: 4 |
		eventType, _ := res.GetStringByName(0, "Event_type")
		this.Ui.Outputf("%8s: %s", "Event", eventType)
		info, _ := res.GetStringByName(0, "Info")
		this.Ui.Outputf("%8s: %s", "Info", info)
		endPos, _ := res.GetIntByName(0, "End_log_pos")
		this.Ui.Outputf("%8s: %d", "NextPos", endPos)

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
