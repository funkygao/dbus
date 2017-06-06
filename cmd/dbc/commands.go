package main

import (
	"os"

	"github.com/funkygao/dbus/cmd/dbc/command"
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
		"checkpoint": func() (cli.Command, error) {
			return &command.Checkpoint{
				Ui:  ui,
				Cmd: cmd,
			}, nil
		},

		"rebalance": func() (cli.Command, error) {
			return &command.Rebalance{
				Ui:  ui,
				Cmd: cmd,
			}, nil
		},

		"config": func() (cli.Command, error) {
			return &command.Config{
				Ui:  ui,
				Cmd: cmd,
			}, nil
		},

		"plugins": func() (cli.Command, error) {
			return &command.Plugins{
				Ui:  ui,
				Cmd: cmd,
			}, nil
		},

		"peek": func() (cli.Command, error) {
			return &command.Peek{
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

		"participants": func() (cli.Command, error) {
			return &command.Participants{
				Ui:  ui,
				Cmd: cmd,
			}, nil
		},

		/*
			"migrate": func() (cli.Command, error) {
				return &command.Migrate{
					Ui:  ui,
					Cmd: cmd,
				}, nil
			},

			"pause": func() (cli.Command, error) {
				return &command.Pause{
					Ui:  ui,
					Cmd: cmd,
				}, nil
			},

			"upgrade": func() (cli.Command, error) {
				return &command.Upgrade{
					Ui:  ui,
					Cmd: cmd,
				}, nil
			},
		*/
	}

}
