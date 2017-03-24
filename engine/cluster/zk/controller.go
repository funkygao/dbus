package zk

import (
	"github.com/funkygao/dbus/engine/cluster"
	"github.com/funkygao/zkclient"
)

var (
	_ zkclient.ZkStateListener = &controller{}
	_ zkclient.ZkChildListener = &controller{}
)

type controller struct {
	kb *keyBuilder
	zc *zkclient.Client

	participantID string

	resources map[string]struct{}
}

// New creates a Controller with zookeeper as underlying storage.
func New(zkSvr string) (cluster.Controller, error) {
	c := &controller{
		kb: newKeyBuilder(),
		zc: zkclient.New(zkSvr),
	}
	if err := c.connectToZookeeper(zkSvr); err != nil {
		return nil, err
	}

	for _, path := range c.kb.persistentKeys() {
		if err := c.zc.CreateEmptyPersistent(path); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *controller) Start() (err error) {

	return
}

func (c *controller) Close() (err error) {
	c.zc.Delete(c.kb.participant(c.participantID))

	c.zc.Disconnect()

	return
}

func (c *controller) IsLeader() bool {
	return false
}

func (c *controller) RegisterParticipent(participant string, weight int) (err error) {
	c.participantID = participant

	return nil
}
