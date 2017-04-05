package zk

import (
	"github.com/funkygao/dbus"
	"github.com/funkygao/zkclient"
)

var (
	_ zkclient.ZkDataListener = &upgrader{}
)

type upgrader struct {
	ctx *controller

	evtCh chan struct{}
}

func newUpgrader(ctx *controller) *upgrader {
	return &upgrader{
		ctx:   ctx,
		evtCh: make(chan struct{}),
	}
}

func (u *upgrader) events() <-chan struct{} {
	return u.evtCh
}

func (u *upgrader) startup() {
	u.ctx.zc.SubscribeDataChanges(u.ctx.kb.upgrade(), u)
}

func (u *upgrader) close() {
	u.ctx.zc.UnsubscribeDataChanges(u.ctx.kb.upgrade(), u)
}

func (u *upgrader) HandleDataChange(dataPath string, lastData []byte) error {
	b, err := u.ctx.zc.Get(u.ctx.kb.upgrade())
	if err != nil {
		return err
	}

	newRevision := string(b)
	if newRevision != dbus.Revision {
		// trigger upgrade
		// if client does not receive this event, will block forever, ok
		u.evtCh <- struct{}{}
	}

	return nil
}

func (u *upgrader) HandleDataDeleted(dataPath string) error {
	return nil
}
