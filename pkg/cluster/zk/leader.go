package zk

import (
	"strconv"

	"github.com/funkygao/dbus/pkg/cluster"
	log "github.com/funkygao/log4go"
	"github.com/funkygao/zkclient"
)

type leader struct {
	ctx *controller

	lastDecision cluster.Decision

	epoch          int
	epochZkVersion int32

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

func (l *leader) fetchEpoch() {
	data, stat, err := l.ctx.zc.GetWithStat(l.ctx.kb.leaderEpoch())
	if err != nil {
		if !zkclient.IsErrNoNode(err) {
			log.Error("%v", err)
		}

		return
	}

	l.epoch, _ = strconv.Atoi(string(data))
	l.epochZkVersion = stat.Version
}

func (l *leader) incrementEpoch() (ok bool) {
	newEpoch := l.epoch + 1
	data := []byte(strconv.Itoa(newEpoch))
	// CAS
	newStat, err := l.ctx.zc.SetWithVersion(l.ctx.kb.leaderEpoch(), data, l.epochZkVersion)
	if err != nil {
		switch {
		case zkclient.IsErrNoNode(err):
			// if path doesn't exist, this is the first controller whose epoch should be 1
			// the following call can still fail if another controller gets elected between checking if the path exists and
			// trying to create the controller epoch path
			if err := l.ctx.zc.CreatePersistent(l.ctx.kb.leaderEpoch(), data); err != nil {
				if zkclient.IsErrNodeExists(err) {
					log.Warn("leader moved to another participant! abort rebalance")
					return
				}

				// unexpected zk err
				log.Error("Error while incrementing controller epoch: %v", err)
				return
			}

			// will go to ok

		case zkclient.IsErrVersionConflict(err):
			log.Warn("leader moved to another participant! abort rebalance")
			return

		default:
			// unexpected zk err
			log.Error("Error while incrementing controller epoch: %v", err)
			return
		}
	}

	ok = true
	l.epoch = newEpoch
	l.epochZkVersion = newStat.Version
	return
}

func (l *leader) onResigningAsLeader() {
	l.ctx.zc.UnsubscribeChildChanges(l.ctx.kb.participants(), l.pcl)
	l.ctx.zc.UnsubscribeChildChanges(l.ctx.kb.resources(), l.rcl)

	l.lastDecision = nil
	l.ctx.elector.leaderID = ""
	l.epoch = 0
	l.epochZkVersion = 0

	log.Trace("[%s] resigned as leader", l.ctx.participant)
}

func (l *leader) onBecomingLeader() {
	l.ctx.zc.SubscribeChildChanges(l.ctx.kb.participants(), l.pcl)
	l.ctx.zc.SubscribeChildChanges(l.ctx.kb.resources(), l.rcl)

	l.fetchEpoch()
	if !l.incrementEpoch() {
		return
	}

	log.Trace("[%s] become controller leader and trigger rebalance!", l.ctx.participant)
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

		// WAL
		walFailure := false
		for participant, resources := range newDecision {
			for _, resource := range resources {
				rs := cluster.NewResourceState()
				rs.LeaderEpoch = l.epoch
				rs.Owner = participant.Endpoint
				// TODO add random sleep here to test race condition
				if err := l.ctx.zc.Set(l.ctx.kb.resourceState(resource.Name), rs.Marshal()); err != nil {
					// zk conn lost? timeout?
					// TODO
					log.Critical("[%s] %s %v", l.ctx.participant, resource.Name, err)
					walFailure = true
					break
				}
			}

			if walFailure {
				break
			}
		}

		l.ctx.onRebalance(l.epoch, newDecision)
	} else {
		log.Trace("[%s] decision stay unchanged, quit rebalance", l.ctx.participant)
	}
}
