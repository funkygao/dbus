package zk

import (
	"github.com/funkygao/dbus/engine/cluster"
	"github.com/funkygao/go-zookeeper/zk"
	log "github.com/funkygao/log4go"
	"github.com/funkygao/zkclient"
)

var (
	_ zkclient.ZkStateListener = &controller{}
)

type controller struct {
	kb *keyBuilder
	zc *zkclient.Client

	participantID string
	weight        int

	leaderID string

	pcl zkclient.ZkChildListener
	lcl zkclient.ZkDataListener

	resources []string

	// only when participant is leader will this callback be triggered.
	rebalanceCallback func(decision map[string][]string)
}

// New creates a Controller with zookeeper as underlying storage.
func New(zkSvr string, participantID string, weight int, callback func(decision map[string][]string)) cluster.Controller {
	return &controller{
		kb:                newKeyBuilder(),
		participantID:     participantID,
		weight:            weight,
		rebalanceCallback: callback,
		zc:                zkclient.New(zkSvr, zkclient.WithWrapErrorWithPath()),
	}
}

func (c *controller) connectToZookeeper() (err error) {
	log.Debug("connecting to zookeeper...")
	c.zc.SubscribeStateChanges(c)

	if err = c.zc.Connect(); err != nil {
		return
	}

	for retries := 0; retries < 3; retries++ {
		if err = c.zc.WaitUntilConnected(c.zc.SessionTimeout()); err == nil {
			log.Debug("connected to zookeeper")
			break
		}

		log.Warn("retry=%d %v", retries, err)
	}

	return
}

func (c *controller) RegisterResources(resources []string) {
	c.resources = resources
	if c.IsLeader() {
		c.rebalance()
	}
}

func (c *controller) Start() (err error) {
	if c.rebalanceCallback == nil {
		return cluster.ErrInvalidCallback
	}

	c.lcl = newLeaderChangeListener(c)
	c.pcl = newParticipantChangeListener(c)

	if err = c.connectToZookeeper(); err != nil {
		return
	}

	for _, path := range c.kb.persistentKeys() {
		if err = c.zc.CreateEmptyPersistent(path); err != nil && err != zk.ErrNodeExists {
			return
		} else {
			err = nil
		}
	}

	c.tryElect()
	return
}

func (c *controller) Close() (err error) {
	if len(c.resources) == 0 {
		return nil
	}

	c.zc.Delete(c.kb.participant(c.participantID))
	c.zc.Disconnect()
	return
}

func (c *controller) IsLeader() bool {
	return c.leaderID == c.participantID
}
