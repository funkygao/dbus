package zk

import (
	"github.com/funkygao/dbus/pkg/cluster"
	"github.com/funkygao/zkclient"
)

// NewManager creates a Manager with zookeeper as underlying storage.
func NewManager(zkSvr string) cluster.Manager {
	return &controller{
		zc: zkclient.New(zkSvr, zkclient.WithWrapErrorWithPath()),
	}
}

func (c *controller) Open() error {
	return c.connectToZookeeper()
}

func (c *controller) Close() {
	c.zc.Disconnect()
}

func (c *controller) RegisterResource(resource cluster.Resource) (err error) {
	if err = c.zc.CreatePersistent(c.kb.resource(resource.Name), resource.Marshal()); err != nil {
		return
	}

	err = c.zc.CreatePersistent(c.kb.resourceState(resource.Name), nil)
	return
}

func (c *controller) RegisteredResources() ([]cluster.Resource, error) {
	resources, marshalled, err := c.zc.ChildrenValues(c.kb.resources())
	if err != nil {
		return nil, err
	}

	liveParticipants := make(map[string]struct{})
	if ps, err := c.LiveParticipants(); err == nil {
		for _, p := range ps {
			liveParticipants[p.Endpoint] = struct{}{}
		}
	}

	r := make([]cluster.Resource, 0)
	for i := range resources {
		res := cluster.Resource{}
		res.From(marshalled[i])
		res.State = cluster.NewResourceState()
		state, err := c.zc.Get(c.kb.resourceState(res.Name))
		if err != nil {
			return r, err
		}
		res.State.From(state)
		if _, present := liveParticipants[res.State.Owner]; !present {
			// its owner has died
			res.State.BecomeOrphan()
		}

		r = append(r, res)
	}

	return r, nil
}

func (c *controller) Leader() (cluster.Participant, error) {
	var p cluster.Participant
	data, err := c.zc.Get(c.kb.leader())
	if err != nil {
		if zkclient.IsErrNoNode(err) {
			return p, cluster.ErrNoLeader
		}

		return p, err
	}

	p.From(data)
	return p, nil
}

func (c *controller) LiveParticipants() ([]cluster.Participant, error) {
	participants, marshalled, err := c.zc.ChildrenValues(c.kb.participants())
	if err != nil {
		return nil, err
	}

	r := make([]cluster.Participant, 0)
	for i := range participants {
		model := cluster.Participant{}
		model.From(marshalled[i])

		r = append(r, model)
	}

	return r, nil
}

func (c *controller) Rebalance() error {
	return c.zc.Delete(c.kb.leader())
}
