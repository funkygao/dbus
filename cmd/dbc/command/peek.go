package command

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/pkg/model"
	"github.com/funkygao/dbus/pkg/myslave"
	"github.com/funkygao/gafka/diagnostics/agent"
	"github.com/funkygao/gocli"
)

type Peek struct {
	Ui  cli.Ui
	Cmd string
}

func (this *Peek) Run(args []string) (exitCode int) {
	var (
		dsn     string
		verbose bool
	)
	cmdFlags := flag.NewFlagSet("peek", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&dsn, "dsn", "", "")
	cmdFlags.BoolVar(&verbose, "verbose", false, "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	agent.HttpAddr = ":10129"
	this.Ui.Infof("pprof agent ready on %s", agent.Start())
	go func() {
		this.Ui.Errorf("%s", <-agent.Errors)
	}()

	e := engine.New(nil)
	e.LoadFrom("")

	slave := myslave.New("peek", dsn, "")
	slave.LoadConfig(e.Conf)
	ready := make(chan struct{})
	go slave.StartReplication(ready)
	<-ready

	tick := time.NewTicker(time.Second * 5)
	rows := slave.Events()
	replErrors := slave.Errors()
	var n, lastN int64
	var row *model.RowsEvent
	for {
		select {
		case err := <-replErrors:
			this.Ui.Error(err.Error())
			return

		case row = <-rows:
			n++
			if verbose {
				this.Ui.Outputf("%+v", row)
			}

		case <-tick.C:
			this.Ui.Infof("%d tps, %s", (n-lastN)/5, time.Unix(int64(row.Timestamp), 0))
			lastN = n
		}
	}
	return
}

func (*Peek) Synopsis() string {
	return "Peek mysql binlog stream from head"
}

func (this *Peek) Help() string {
	help := fmt.Sprintf(`
Usage: %s peek [options]

    %s

Options:

    -dsn dsn
     Output of dbc resources

    -verbose

`, this.Cmd, this.Synopsis())
	return strings.TrimSpace(help)
}
