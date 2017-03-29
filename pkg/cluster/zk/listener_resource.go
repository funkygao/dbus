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
	if !r.ctx.amLeader() {
		log.Warn("[%s] was leader but resigned now", r.ctx.participant)
		return nil
	}

	log.Trace("[%s] resources changed, trigger rebalance", r.ctx.participant)
	r.ctx.doRebalance()
	return nil
}
