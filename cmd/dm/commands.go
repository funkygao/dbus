package main

import (
	"os"

	"github.com/funkygao/dbus/cmd/dm/command"
	"github.com/funkygao/gocli"
)

var commands map[string]cli.CommandFactory

func init() {
	ui := &cli.ColoredUi{
		Ui: &cli.BasicUi{
			Writer:      os.Stdout,
			Reader:      os.Stdin,
			ErrorWriter: os.Stderr,
		},
		OutputColor: cli.UiColorNone,
		InfoColor:   cli.UiColorGreen,
		ErrorColor:  cli.UiColorRed,
		WarnColor:   cli.UiColorYellow,
	}
	cmd := os.Args[0]

	commands = map[string]cli.CommandFactory{
		"status": func() (cli.Command, error) {
			return &command.Status{
				Ui:  ui,
				Cmd: cmd,
			}, nil
		},

		"zones": func() (cli.Command, error) {
			return &command.Zones{
				Ui:  ui,
				Cmd: cmd,
			}, nil
		},

		"clusters": func() (cli.Command, error) {
			return &command.Clusters{
				Ui:  ui,
				Cmd: cmd,
			}, nil
		},

		"resources": func() (cli.Command, error) {
			return &command.Resources{
				Ui:  ui,
				Cmd: cmd,
			}, nil
		},
	}

}
