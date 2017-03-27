package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/funkygao/gocli"
)

type Migrate struct {
	Ui  cli.Ui
	Cmd string
}

func (this *Migrate) Run(args []string) (exitCode int) {
	cmdFlags := flag.NewFlagSet("migrate", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	this.showWorkflow()

	return
}

func (this *Migrate) showWorkflow() {
	this.Ui.Outputf("option A:")
	this.Ui.Outputf("    gh-ost --database a --table a --alter 'add bb int;' --user root --debug --allow-on-master --initially-drop-ghost-table --execute")
	this.Ui.Output("")
	this.Ui.Outputf("option B:")
	this.Ui.Outputf("    mysqldump --host={host} --port={port} --user={user} --password={pass} --master-data --single-transaction --skip-lock-tables --compact --quick --skip-opt --no-create-info --skip-extended-insert --all-databases > d.sql")
	this.Ui.Outputf("    locate 'CHANGE MASTER TO MASTER_LOG_FILE={name}, MASTER_LOG_POS={pos}' in d.sql")
	this.Ui.Outputf("    import on target mysql instance: mysql < d.sql")
	this.Ui.Outputf("    setup position for dbusd by {name}:{pos}, startup dbusd")

}

func (*Migrate) Synopsis() string {
	return "Workflow of migrating a mysql database"
}

func (this *Migrate) Help() string {
	help := fmt.Sprintf(`
Usage: %s migrate [options]

    %s

`, this.Cmd, this.Synopsis())
	return strings.TrimSpace(help)
}
