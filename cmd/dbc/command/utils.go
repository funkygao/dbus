package command

import (
	"github.com/funkygao/dbus/pkg/cluster"
	czk "github.com/funkygao/dbus/pkg/cluster/zk"
	"github.com/funkygao/gafka/ctx"
)

func openClusterManager(zone string) cluster.Manager {
	mgr := czk.NewManager(ctx.ZoneZkAddrs(zone))
	swallow(mgr.Open())

	return mgr
}

func swallow(err error) {
	if err != nil {
		panic(err)
	}
}
