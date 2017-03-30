package zk

import (
	"github.com/funkygao/dbus/pkg/cluster"
)

func (c *controller) Open() error {
	return c.connectToZookeeper()
}

func (c *controller) Close() {
	c.zc.Disconnect()
}

func (c *controller) RegisterResource(resource cluster.Resource) error {
	return c.zc.CreatePersistent(c.kb.resource(resource.Name), resource.Marshal())
}

func (c *controller) RegisteredResources() ([]cluster.Resource, error) {
	resources, marshalled, err := c.zc.ChildrenValues(c.kb.resources())
	if err != nil {
		return nil, err
	}

	r := make([]cluster.Resource, 0)
	for i, _ := range resources {
		model := cluster.Resource{}
		model.From(marshalled[i])

		r = append(r, model)
	}

	return r, nil
}

func (c *controller) Controller() (cluster.Participant, error) {
	var p cluster.Participant
	data, err := c.zc.Get(c.kb.controller())
	if err != nil {
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
	for i, _ := range participants {
		model := cluster.Participant{}
		model.From(marshalled[i])

		r = append(r, model)
	}

	return r, nil
}
