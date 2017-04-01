package zk

import (
	"github.com/funkygao/dbus/pkg/cluster"
	log "github.com/funkygao/log4go"
	"github.com/funkygao/zkclient"
)

type leader struct {
	ctx *controller

	lastDecision cluster.Decision

	pcl zkclient.ZkChildListener // leader watches live participants
	rcl zkclient.ZkChildListener // leader watches resources
}

func newLeader(ctx *controller) *leader {
	return &leader{
		ctx: ctx,
		pcl: newParticipantChangeListener(ctx),
		rcl: newResourceChangeListener(ctx),
	}
}

func (l *leader) onResigningAsLeader() {
	l.ctx.zc.UnsubscribeChildChanges(l.ctx.kb.participants(), l.pcl)
	l.ctx.zc.UnsubscribeChildChanges(l.ctx.kb.resources(), l.rcl)
	l.lastDecision = nil
	l.ctx.leaderID = ""
}

func (l *leader) onBecomingLeader() {
	l.ctx.zc.SubscribeChildChanges(l.ctx.kb.participants(), l.pcl)
	l.ctx.zc.SubscribeChildChanges(l.ctx.kb.resources(), l.rcl)

	log.Trace("become controller leader and trigger rebalance!")
	l.doRebalance()
}

/// rebalance happens on controller leader when:
// 1. participants change
// 2. resources change
// 3. becoming leader
func (l *leader) doRebalance() {
	participants, err := l.ctx.LiveParticipants()
	if err != nil {
		// TODO
		log.Critical("[%s] %s", l.ctx.participant, err)
		return
	}
	if len(participants) == 0 {
		log.Critical("[%s] no alive participants found", l.ctx.participant)
		return
	}

	resources, err := l.ctx.RegisteredResources()
	if err != nil {
		// TODO
		log.Critical("[%s] %s", l.ctx.participant, err)
		return
	}

	newDecision := l.ctx.strategyFunc(participants, resources)
	if !newDecision.Equals(l.lastDecision) {
		l.lastDecision = newDecision

		l.ctx.onRebalance(newDecision)
	} else {
		log.Trace("[%s] decision stay unchanged, quit rebalance", l.ctx.participant)
	}
}
