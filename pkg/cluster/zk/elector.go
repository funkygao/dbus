package zk

import (
	log "github.com/funkygao/log4go"
	"github.com/funkygao/zkclient"
)

var (
	_ zkclient.ZkDataListener = &leaderElector{}
)

type leaderElector struct {
	*controller

	leaderID string // participant id of the leader

	onResigningAsLeader func()
	onBecomingLeader    func()
}

func newLeaderElector(ctx *controller, onBecomingLeader func(), onResigningAsLeader func()) *leaderElector {
	return &leaderElector{
		controller:          ctx,
		onBecomingLeader:    onBecomingLeader,
		onResigningAsLeader: onResigningAsLeader,
	}
}

func (l *leaderElector) startup() {
	// watch for leader changes
	l.zc.SubscribeDataChanges(l.kb.controller(), l)
	l.elect()
}

func (l *leaderElector) fetchLeaderID() string {
	b, err := l.zc.Get(l.kb.controller())
	if err != nil {
		return ""
	}

	return string(b)
}

func (l *leaderElector) elect() (win bool) {
	log.Trace("[%s] elect...", l.participant)

	// we can get here during the initial startup and the HandleDataDeleted callback.
	// because of the potential race condition, it's possible that the leader has already
	// been elected when we get here.
	l.leaderID = l.fetchLeaderID()
	if l.leaderID != "" {
		log.Trace("[%s] found leader: %s", l.participant, l.leaderID)
		return
	}

	if err := l.zc.CreateLiveNode(l.kb.controller(), l.participant.Marshal(), 2); err == nil {
		log.Trace("[%s] elect win!", l.participant)

		win = true
		l.leaderID = l.participant.Endpoint
		l.onBecomingLeader()
	} else {
		log.Trace("[%s] lose win :-)", l.participant)

		l.leaderID = l.fetchLeaderID() // refresh
		if l.leaderID == "" {
			log.Warn("[%s] a leader has been elected but just resigned, this will lead to another round of election", l.participant)
		}
	}

	return
}

func (l *leaderElector) close() {
	// needn't delete /controller znode because when
	// zkclient closes the ephemeral znode will disappear automatically
	l.leaderID = ""
}

func (l *leaderElector) amLeader() bool {
	return l.leaderID == l.participant.Endpoint
}

func (l *leaderElector) HandleDataChange(dataPath string, lastData []byte) error {
	l.leaderID = l.fetchLeaderID()
	log.Trace("[%s] new leader is %s", l.participant, l.leaderID)
	return nil
}

func (l *leaderElector) HandleDataDeleted(dataPath string) error {
	log.Trace("[%s] leader[%s] gone!", l.participant, l.leaderID)

	if l.amLeader() {
		l.onResigningAsLeader()
	}

	l.elect()
	return nil
}
