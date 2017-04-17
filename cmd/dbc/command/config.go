package command

import (
	"flag"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gafka/zk"
	"github.com/funkygao/gocli"
)

type Config struct {
	Ui  cli.Ui
	Cmd string
}

func (this *Config) Run(args []string) (exitCode int) {
	var (
		zone     string
		fromFile string
	)
	cmdFlags := flag.NewFlagSet("config", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&zone, "z", ctx.ZkDefaultZone(), "")
	cmdFlags.StringVar(&fromFile, "from", "", "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	zkzone := zk.NewZkZone(zk.DefaultConfig(zone, ctx.ZoneZkAddrs(zone)))
	path := "/dbus/conf"
	if fromFile != "" {
		data, err := ioutil.ReadFile(fromFile)
		if err != nil {
			this.Ui.Error(err.Error())
			return
		}
		if _, err := zkzone.Conn().Set(path, data, -1); err != nil {
			this.Ui.Error(err.Error())
		} else {
			this.Ui.Info("ok")
		}

		return
	}

	// display configuration
	data, _, err := zkzone.Conn().Get(path)
	if err != nil {
		this.Ui.Error(err.Error())
	} else {
		this.Ui.Output(string(data))
	}
	return
}

func (*Config) Synopsis() string {
	return "Setup central configuration in zookeeper"
}

func (this *Config) Help() string {
	help := fmt.Sprintf(`
Usage: %s config [options]

    %s

Options:

    -z zone

    -from filename
      Import to central config from local file.

`, this.Cmd, this.Synopsis())
	return strings.TrimSpace(help)
}
