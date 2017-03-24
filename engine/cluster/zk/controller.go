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

	resources map[string]struct{}
}

// New creates a Controller with zookeeper as underlying storage.
func New(zkSvr string, participantID string, weight int) (cluster.Controller, error) {
	c := &controller{
		kb:            newKeyBuilder(),
		participantID: participantID,
		weight:        weight,
		zc:            zkclient.New(zkSvr, zkclient.WithWrapErrorWithPath()),
	}
	c.lcl = newLeaderChangeListener(c)
	c.pcl = newParticipantChangeListener(c)
	if err := c.connectToZookeeper(zkSvr); err != nil {
		return nil, err
	}

	for _, path := range c.kb.persistentKeys() {
		if err := c.zc.CreateEmptyPersistent(path); err != nil && err != zk.ErrNodeExists {
			return nil, err
		}
	}

	return c, nil
}

func (c *controller) connectToZookeeper(zkSvr string) (err error) {
	log.Debug("connecting to %s...", zkSvr)
	c.zc.SubscribeStateChanges(c)

	if err = c.zc.Connect(); err != nil {
		return
	}

	for retries := 0; retries < 3; retries++ {
		if err = c.zc.WaitUntilConnected(c.zc.SessionTimeout()); err == nil {
			log.Debug("connected to %s", zkSvr)
			break
		}

		log.Warn("%s retry=%d %v", zkSvr, retries, err)
	}

	return
}

func (c *controller) Start() (err error) {
	c.tryElect()
	return
}

func (c *controller) Close() (err error) {
	c.zc.Delete(c.kb.participant(c.participantID))

	c.zc.Disconnect()
	return
}

func (c *controller) IsLeader() bool {
	return c.leaderID == c.participantID
}
