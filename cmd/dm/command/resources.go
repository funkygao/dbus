package command

import (
	"fmt"
	"strings"

	"github.com/funkygao/gocli"
)

type Resources struct {
	Ui  cli.Ui
	Cmd string
}

func (this *Resources) Run(args []string) (exitCode int) {
	return
}

func (*Resources) Synopsis() string {
	return "Manipulate helix resources in a cluster"
}

func (this *Resources) Help() string {
	help := fmt.Sprintf(`
Usage: %s clusters [options]

    %s

Options:

    -z zone

`, this.Cmd, this.Synopsis())
	return strings.TrimSpace(help)
}
