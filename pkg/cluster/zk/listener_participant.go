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
	if !p.ctx.amLeader() {
		log.Warn("[%s] was leader but resigned now", p.ctx.participant)
		return nil
	}

	log.Trace("[%s] participants changed, trigger rebalance", p.ctx.participant)
	p.ctx.doRebalance()
	return nil
}
