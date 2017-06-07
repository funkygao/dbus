package command

import (
	"flag"
	"fmt"
	"io/ioutil"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gafka/zk"
	"github.com/funkygao/gocli"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// TODO github.com/yudai/gojsondiff
type Config struct {
	Ui  cli.Ui
	Cmd string
}

func (this *Config) Run(args []string) (exitCode int) {
	var (
		zone     string
		fromFile string
		cluster  string
		diff     string
		vers     bool
	)
	cmdFlags := flag.NewFlagSet("config", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&zone, "z", ctx.ZkDefaultZone(), "")
	cmdFlags.StringVar(&cluster, "c", "", "")
	cmdFlags.StringVar(&fromFile, "from", "", "")
	cmdFlags.BoolVar(&vers, "vers", false, "")
	cmdFlags.StringVar(&diff, "diff", "", "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	zkzone := zk.NewZkZone(zk.DefaultConfig(zone, ctx.ZoneZkAddrs(zone)))
	if len(cluster) == 0 {
		if cluster = zkzone.DefaultDbusCluster(); cluster == "" {
			this.Ui.Error("-c required")
			return
		}
	}

	switch {
	case fromFile != "":
		this.importFromFile(zkzone, cluster, fromFile)

	case diff != "":
		tuples := strings.SplitN(diff, ":", 2)
		if len(tuples) != 2 {
			this.Ui.Output(this.Help())
			return
		}
		this.compareVers(zkzone, cluster, tuples[0], tuples[1])

	case vers:
		this.listVers(zkzone, cluster)

	default:
		// display configuration
		data, _, err := zkzone.Conn().Get(zk.DbusConfig(cluster))
		if err != nil {
			this.Ui.Error(err.Error())
		} else {
			this.Ui.Output(string(data))
		}
	}

	return
}

func (this *Config) importFromFile(zkzone *zk.ZkZone, cluster string, fromFile string) {
	data, err := ioutil.ReadFile(fromFile)
	if err != nil {
		this.Ui.Error(err.Error())
		return
	}

	if zkData, _, err := zkzone.Conn().Get(zk.DbusConfig(cluster)); err == nil {
		if strings.TrimSpace(string(data)) == strings.TrimSpace(string(zkData)) {
			this.Ui.Warn("config same as inside zk, import gave up")
			return
		}
	}

	if _, err := zkzone.Conn().Set(zk.DbusConfig(cluster), data, -1); err != nil {
		this.Ui.Error(err.Error())
	}

	// generate the history version
	vers, _, err := zkzone.Conn().Children(zk.DbusConfigDir(cluster))
	swallow(err)

	var maxVer int
	if len(vers) > 0 {
		sort.Strings(vers)
		maxVer, err = strconv.Atoi(vers[len(vers)-1])
		if err != nil {
			this.Ui.Errorf("%+v", vers)
			return
		}
	}

	hisVerPath := path.Join(zk.DbusConfigDir(cluster), strconv.Itoa(maxVer+1))
	if err = zkzone.CreatePermenantZnode(hisVerPath, data); err != nil {
		this.Ui.Error(err.Error())
		return
	}

	this.Ui.Infof("ok with ver:%d", maxVer+1)
}

func (this *Config) listVers(zkzone *zk.ZkZone, cluster string) {
	vers, _, err := zkzone.Conn().Children(zk.DbusConfigDir(cluster))
	if err != nil {
		this.Ui.Error(err.Error())
		return
	}

	sort.Strings(vers)
	for _, v := range vers {
		this.Ui.Output(v)
	}
}

func (this *Config) compareVers(zkzone *zk.ZkZone, cluster, v1, v2 string) {
	data1, _, err := zkzone.Conn().Get(path.Join(zk.DbusConfigDir(cluster), v1))
	if err != nil {
		this.Ui.Error(err.Error())
		return
	}

	data2, _, err := zkzone.Conn().Get(path.Join(zk.DbusConfigDir(cluster), v2))
	if err != nil {
		this.Ui.Error(err.Error())
		return
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(data1), string(data2), false)
	this.Ui.Output(dmp.DiffPrettyText(diffs))
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

    -c cluster

    -from filename
      Import to central config from local file.

    -vers
      List all versions of config.

    -diff v1:v2
      Display differences between two versions.

`, this.Cmd, this.Synopsis())
	return strings.TrimSpace(help)
}
