package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/funkygao/gocli"
	"github.com/github/gh-ost/go/base"
	"github.com/github/gh-ost/go/logic"
	"github.com/outbrain/golib/log"
)

type Migrate struct {
	Ui  cli.Ui
	Cmd string
}

func (this *Migrate) Run(args []string) (exitCode int) {
	ctx := base.GetMigrationContext()
	cmdFlags := flag.NewFlagSet("migrate", flag.ContinueOnError)
	cmdFlags.StringVar(&ctx.InspectorConnectionConfig.Key.Hostname, "host", "127.0.0.1", "")
	cmdFlags.IntVar(&ctx.InspectorConnectionConfig.Key.Port, "port", 3306, "")
	cmdFlags.StringVar(&ctx.DatabaseName, "db", "", "")
	cmdFlags.StringVar(&ctx.OriginalTableName, "table", "", "")
	cmdFlags.StringVar(&ctx.AlterStatement, "alter", "", "")
	cmdFlags.StringVar(&ctx.CliUser, "user", "root", "")
	cmdFlags.StringVar(&ctx.CliPassword, "pass", "", "")
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	ctx.AllowedRunningOnMaster = true
	ctx.OkToDropTable = true
	ctx.InitiallyDropOldTable = true
	ctx.InitiallyDropGhostTable = true
	ctx.ServeSocketFile = "/tmp/dbc.migrate.sock"
	ctx.ApplyCredentials()
	log.EnableSyslogWriter("")
	log.SetLevel(log.DEBUG)

	log.Infof("starting")
	migrator := logic.NewMigrator()
	if err := migrator.Migrate(); err != nil {
		migrator.ExecOnFailureHook()
		log.Fatale(err)
	}
	log.Info("OK")

	return
}

func (*Migrate) Synopsis() string {
	return "Asynchronous MySQL online schema migration tool"
}

func (this *Migrate) Help() string {
	help := fmt.Sprintf(`
Usage: %s migrate [options]

    %s

Options:

    -user db username

    -pass db password

    -db dabase

    -table table

    -alter statement
      e,g.
      dbc migrate -db test -table x -alter 'add age int'


`, this.Cmd, this.Synopsis())
	return strings.TrimSpace(help)
}
