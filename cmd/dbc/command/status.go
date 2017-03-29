package command

import (
	"fmt"
	"strings"

	"github.com/funkygao/gocli"
)

type Status struct {
	Ui  cli.Ui
	Cmd string
}

func (this *Status) Run(args []string) (exitCode int) {
	return
}

func (*Status) Synopsis() string {
	return "Display current status of dbus"
}

func (this *Status) Help() string {
	help := fmt.Sprintf(`
Usage: %s status [options]

    %s

Options:

    -z zone

`, this.Cmd, this.Synopsis())
	return strings.TrimSpace(help)
}
