package zk

import (
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

func (p *resourceChangeListener) HandleChildChange(parentPath string, lastChilds []string) error {
	p.ctx.rebalance()
	return nil
}
