package zk

import (
	log "github.com/funkygao/log4go"
	"github.com/funkygao/zkclient"
)

var (
	_ zkclient.ZkChildListener = &resourceChangeListener{}
)

type resourceChangeListener struct {
	ctx *controller
}

func newResourceChangeListener(ctx *controller) *resourceChangeListener {
	return &resourceChangeListener{ctx: ctx}
}

func (r *resourceChangeListener) HandleChildChange(parentPath string, lastChilds []string) error {
	log.Trace("[%s] resources changed, trigger rebalance", r.ctx.participantID)
	r.ctx.rebalance()
	return nil
}
