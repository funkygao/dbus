package zk

import (
	"github.com/funkygao/dbus/engine/cluster"
)

type controller struct {
}

func New(zkSvr string) cluster.Controller {
	return &controller{}
}

func (c *controller) Start() error {
	return nil
}

func (c *controller) Close() error {
	return nil
}

func (c *controller) WaitForTicket(participant string) error {
	return nil
}

func (c *controller) RegisterResource(resource string) error {
	return nil
}

func (c *controller) RegisterParticipent(participant string) error {
	return nil
}
