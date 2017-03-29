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
	for i, encodedResource := range resources {
		model := cluster.Resource{}
		model.From(marshalled[i])
		res, _ := c.kb.decodeResource(encodedResource)
		model.Name = res

		r = append(r, model)
	}

	return r, nil
}

func (c *controller) LiveParticipants() ([]cluster.Participant, error) {
	participants, marshalled, err := c.zc.ChildrenValues(c.kb.participants())
	if err != nil {
		return nil, err
	}

	r := make([]cluster.Participant, 0)
	for i, participant := range participants {
		model := cluster.Participant{}
		model.From(marshalled[i])
		model.Endpoint = participant

		r = append(r, model)
	}

	return r, nil
}
