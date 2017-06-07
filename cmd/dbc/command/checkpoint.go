package command

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/funkygao/columnize"
	"github.com/funkygao/dbus/pkg/checkpoint"
	czk "github.com/funkygao/dbus/pkg/checkpoint/store/zk"
	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gafka/zk"
	"github.com/funkygao/gocli"
)

type Checkpoint struct {
	Ui  cli.Ui
	Cmd string
}

func (this *Checkpoint) Run(args []string) (exitCode int) {
	var (
		zone    string
		cluster string
		topMode bool
	)
	cmdFlags := flag.NewFlagSet("checkpoint", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&zone, "z", ctx.ZkDefaultZone(), "")
	cmdFlags.StringVar(&cluster, "c", "", "")
	cmdFlags.BoolVar(&topMode, "top", false, "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	zkzone := zk.NewZkZone(zk.DefaultConfig(zone, ctx.ZoneZkAddrs(zone)))
	mgr := czk.NewManager(zkzone)

	lastStates := make(map[string]checkpoint.State)
	for {
		states, err := mgr.AllStates()
		if err != nil {
			this.Ui.Error(err.Error())
			return 2
		}

		lines := []string{"Input|Scheme|DSN|Position"}
		if topMode {
			lines = []string{"Input|Scheme|DSN|Position|Delta"}
		}
		for _, state := range states {
			if topMode {
				if last, present := lastStates[state.DSN()]; present {
					lines = append(lines, fmt.Sprintf("%s|%s|%s|%s|%s", state.Name(), state.Scheme(), state.DSN(), state.String(), state.Delta(last)))
				} else {
					lines = append(lines, fmt.Sprintf("%s|%s|%s|%s|-", state.Name(), state.Scheme(), state.DSN(), state.String()))
				}
				lastStates[state.DSN()] = state
			} else {
				lines = append(lines, fmt.Sprintf("%s|%s|%s|%s", state.Name(), state.Scheme(), state.DSN(), state.String()))
			}
		}

		if len(lines) > 1 {
			this.Ui.Output(columnize.SimpleFormat(lines))
		}

		if !topMode {
			break
		}

		time.Sleep(time.Second * 3)
		refreshScreen()
	}

	return
}

func (*Checkpoint) Synopsis() string {
	return "Manages cluster checkpoint"
}

func (this *Checkpoint) Help() string {
	help := fmt.Sprintf(`
Usage: %s checkpoint [options]

    %s

Options:

    -z zone

    -c cluster

    -top
      Run in top mode.

`, this.Cmd, this.Synopsis())
	return strings.TrimSpace(help)
}
