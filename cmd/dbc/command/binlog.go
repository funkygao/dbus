package command

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/plugins/input/mysql"
	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gocli"

	// bootstrap plugins
	_ "github.com/funkygao/dbus/plugins/filter"
	_ "github.com/funkygao/dbus/plugins/input"
	_ "github.com/funkygao/dbus/plugins/output"
)

type Binlog struct {
	Ui  cli.Ui
	Cmd string

	zone           string
	fn             string
	plugin         string
	binlogsOfInput string
	binlogPos      string
}

func (this *Binlog) Run(args []string) (exitCode int) {
	cmdFlags := flag.NewFlagSet("binlog", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&this.zone, "z", ctx.ZkDefaultZone(), "")
	cmdFlags.StringVar(&this.fn, "c", "", "")
	cmdFlags.StringVar(&this.plugin, "id", "", "")
	cmdFlags.StringVar(&this.binlogsOfInput, "logs", "", "")
	cmdFlags.StringVar(&this.binlogPos, "pos", "", "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	e := engine.New(nil).LoadFrom(this.fn)
	mgr := openClusterManager(this.zone)
	defer mgr.Close()

	resources, err := mgr.RegisteredResources()
	if err != nil {
		panic(err)
	}

	switch {
	case this.binlogPos != "":
		ir, present := e.InputRunners[this.plugin]
		if !present {
			this.Ui.Errorf("%s not found", this.plugin)
			return 2
		}

		tuples := strings.SplitN(this.binlogPos, "-", 2)
		if len(tuples) != 2 {
			this.Ui.Output(this.Help())
			return 2
		}
		pos, _ := strconv.Atoi(tuples[1])

		my := ir.Plugin().(*mysql.MysqlbinlogInput)
		for _, res := range resources {
			if res.InputPlugin == this.plugin {
				my.ConnectMyslave(res.DSN())
				break
			}
		}
		res, err := my.MySlave().BinlogByPos(tuples[0], pos)
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

	case this.binlogsOfInput != "":
		ir, present := e.InputRunners[this.binlogsOfInput]
		if !present {
			this.Ui.Errorf("%s not found", this.binlogsOfInput)
			return 2
		}

		my := ir.Plugin().(*mysql.MysqlbinlogInput)
		for _, res := range resources {
			if res.InputPlugin == this.binlogsOfInput {
				my.ConnectMyslave(res.DSN())
				break
			}
		}
		logs, err := my.MySlave().MasterBinlogs()
		if err != nil {
			panic(err)
		}

		this.Ui.Info(this.binlogsOfInput)
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
Usage: %s binlog [options]

    %s

    Must run on the same host as dbusd.

Options:

    -c config location
      If empty, load from zookeeper
      zk location example:
      localhost:2181/dbus/conf

    -id input plugin name

    -logs input plugin name
      Display a master binlog status

    -pos file-offset
      Print a single mysql binlog event content
      Hint:
        you can use 'dbc checkpoint' to locate position

`, this.Cmd, this.Synopsis())
	return strings.TrimSpace(help)
}
