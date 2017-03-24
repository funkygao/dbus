package zk

import (
	"github.com/funkygao/zkclient"
)

var (
	_ zkclient.ZkDataListener = &leaderChangeListener{}
)

type leaderChangeListener struct {
	ctx *controller
}

func newLeaderChangeListener(ctx *controller) *leaderChangeListener {
	return &leaderChangeListener{ctx: ctx}
}

func (l *leaderChangeListener) HandleDataChange(dataPath string, lastData []byte) error {
	l.ctx.refreshLeaderID()
	return nil
}

func (l *leaderChangeListener) HandleDataDeleted(dataPath string) error {
	l.ctx.tryElect()
	return nil
}
