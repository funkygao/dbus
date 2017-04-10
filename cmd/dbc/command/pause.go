package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gocli"
)

type Pause struct {
	Ui  cli.Ui
	Cmd string

	zone       string
	inputName  string
	resumeMode bool
}

func (this *Pause) Run(args []string) (exitCode int) {
	cmdFlags := flag.NewFlagSet("pause", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&this.zone, "z", ctx.DefaultZone(), "")
	cmdFlags.StringVar(&this.inputName, "in", "", "")
	cmdFlags.BoolVar(&this.resumeMode, "r", false, "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	if len(this.inputName) == 0 {
		this.Ui.Output(this.Help())
		return 2
	}

	mgr := openClusterManager(this.zone)
	defer mgr.Close()

	var q string
	if this.resumeMode {
		q = fmt.Sprintf("api/v1/resume/%s", this.inputName)
	} else {
		q = fmt.Sprintf("api/v1/pause/%s", this.inputName)
	}
	if err := mgr.CallParticipants("PUT", q); err != nil {
		this.Ui.Error(err.Error())
		return
	}

	this.Ui.Info("ok")
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
