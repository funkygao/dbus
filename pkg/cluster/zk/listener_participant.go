package zk

import (
	"github.com/funkygao/zkclient"
)

var (
	_ zkclient.ZkChildListener = &participantChangeListener{}
)

type participantChangeListener struct {
	ctx *controller
}

func newParticipantChangeListener(ctx *controller) *participantChangeListener {
	return &participantChangeListener{ctx: ctx}
}

func (p *participantChangeListener) HandleChildChange(parentPath string, lastChilds []string) error {
	p.ctx.rebalance()
	return nil
}
