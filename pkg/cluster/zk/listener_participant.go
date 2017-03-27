package zk

import (
	log "github.com/funkygao/log4go"
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
	log.Trace("[%s] participants changed, trigger rebalance", p.ctx.participantID)
	p.ctx.rebalance()
	return nil
}
