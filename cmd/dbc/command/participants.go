package command

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/funkygao/columnize"
	"github.com/funkygao/dbus/pkg/cluster"
	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gafka/zk"
	"github.com/funkygao/gocli"
)

type Participants struct {
	Ui  cli.Ui
	Cmd string

	zone       string
	cluster    string
	showQueues bool
}

func (this *Participants) Run(args []string) (exitCode int) {
	cmdFlags := flag.NewFlagSet("participants", flag.ContinueOnError)
	cmdFlags.Usage = func() { this.Ui.Output(this.Help()) }
	cmdFlags.StringVar(&this.zone, "z", ctx.ZkDefaultZone(), "")
	cmdFlags.StringVar(&this.cluster, "c", "", "")
	cmdFlags.BoolVar(&this.showQueues, "q", false, "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	zkzone := zk.NewZkZone(zk.DefaultConfig(this.zone, ctx.ZoneZkAddrs(this.zone)))
	if len(this.cluster) == 0 {
		if this.cluster = zkzone.DefaultDbusCluster(); this.cluster == "" {
			this.Ui.Error("-c required")
			return
		}
	}

	mgr := openClusterManager(this.zone, this.cluster)
	defer mgr.Close()

	leader, err := mgr.Leader()
	if err != nil {
		this.Ui.Error(err.Error())
		return
	}

	// list all resources
	ps, err := mgr.LiveParticipants()
	if err != nil {
		this.Ui.Error(err.Error())
		return
	}

	if this.showQueues {
		for _, p := range ps {
			queues, errs := callAPI(p, "queues", "GET", "")
			if len(errs) > 0 {
				this.Ui.Errorf("%+v %+v", p, errs)
				return
			}

			if p.Equals(leader) {
				this.Ui.Output(p.Endpoint + "*")
			} else {
				this.Ui.Output(p.Endpoint)
			}
			this.Ui.Output(queues)
		}

		return
	}

	d := cluster.MakeDecision()
	decision, errs := callAPI(leader, "decision", "GET", "")
	if len(errs) > 0 {
		this.Ui.Errorf("%+v", errs)
		return
	}

	swallow(json.Unmarshal([]byte(decision), &d))

	lines := []string{"Endpoint|State|Weight|Revision|Resources"}
	for _, p := range ps {
		if p.Equals(leader) {
			lines = append(lines, fmt.Sprintf("%s*|%s|%d|%s|%+v", p.Endpoint, p.StateText(),
				p.Weight, p.Revision, this.getResources(p, d)))
		} else {
			lines = append(lines, fmt.Sprintf("%s|%s|%d|%s|%+v", p.Endpoint, p.StateText(),
				p.Weight, p.Revision, this.getResources(p, d)))
		}
	}

	if len(lines) > 1 {
		this.Ui.Output(columnize.SimpleFormat(lines))
	}

	return
}

func (*Participants) getResources(p cluster.Participant, d cluster.Decision) []cluster.Resource {
	for participant, rs := range d {
		if p.Endpoint == participant.Endpoint {
			return rs
		}
	}
	return nil
}

func (*Participants) Synopsis() string {
	return "Display live participants in cluster"
}

func (this *Participants) Help() string {
	help := fmt.Sprintf(`
Usage: %s participants [options]

    %s

Options:

    -z zone

    -c cluster

    -q
     Display each participant's internal dataflow queues.

`, this.Cmd, this.Synopsis())
	return strings.TrimSpace(help)
}
