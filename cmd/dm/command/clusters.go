package command

import (
	"fmt"
	"strings"

	"github.com/funkygao/gocli"
)

type Clusters struct {
	Ui  cli.Ui
	Cmd string
}

func (this *Clusters) Run(args []string) (exitCode int) {
	return
}

func (*Clusters) Synopsis() string {
	return "Management of helix clusters"
}

func (this *Clusters) Help() string {
	help := fmt.Sprintf(`
Usage: %s clusters [options]

    %s

Options:

    -z zone   

`, this.Cmd, this.Synopsis())
	return strings.TrimSpace(help)
}
