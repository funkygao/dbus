package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/funkygao/gocli"
)

type Pause struct {
	Ui  cli.Ui
	Cmd string

	inputName  string
	resumeMode bool
}

func (this *Pause) Run(args []string) (exitCode int) {
	cmdFlags := flag.NewFlagSet("pause", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&this.inputName, "in", "", "")
	cmdFlags.BoolVar(&this.resumeMode, "r", false, "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	// call api server
	// /api/v1/pause/{input}
	// /api/v1/resume/{input}

	return
}

func (*Pause) Synopsis() string {
	return "Pause/Resume an Input plugin"
}

func (this *Pause) Help() string {
	help := fmt.Sprintf(`
Usage: %s pause [options]

    %s

Options:

    -z zone

    -in input name

    -r 
      Resume instead of pause.

`, this.Cmd, this.Synopsis())
	return strings.TrimSpace(help)
}
