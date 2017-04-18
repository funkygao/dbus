package zk

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/funkygao/dbus"
	"github.com/funkygao/dbus/pkg/cluster"
	"github.com/funkygao/gorequest"
	"github.com/funkygao/zkclient"
)

// NewManager creates a Manager with zookeeper as underlying storage.
func NewManager(zkSvr string, zroot string) cluster.Manager {
	if len(zroot) > 0 {
		rootPath = zroot
	}

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

func (c *controller) TriggerUpgrade() (err error) {
	data := []byte(dbus.Revision)
	err = c.zc.Set(c.kb.upgrade(), data)
	if zkclient.IsErrNoNode(err) {
		return c.zc.CreatePersistent(c.kb.upgrade(), data)
	}

	return
}

func (c *controller) CurrentDecision() cluster.Decision {
	if !c.amLeader() {
		return nil
	}

	return c.leader.lastDecision
}

func (c *controller) RegisterResource(resource cluster.Resource) (err error) {
	if err = c.zc.CreatePersistent(c.kb.resource(resource.Name), resource.Marshal()); err != nil {
		return
	}

	err = c.zc.CreatePersistent(c.kb.resourceState(resource.Name), nil)
	return
}

func (c *controller) UnregisterResource(resource cluster.Resource) (err error) {
	if err = c.zc.Delete(c.kb.resourceState(resource.Name)); err != nil {
		return
	}

	err = c.zc.Delete(c.kb.resource(resource.Name))
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

	var r []cluster.Resource
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

func (c *controller) CallParticipants(method string, q string) (err error) {
	var ps []cluster.Participant
	ps, err = c.LiveParticipants()
	if err != nil {
		return
	}

	var wg sync.WaitGroup
	for _, p := range ps {
		wg.Add(1)

		targetURI := fmt.Sprintf("%s/%s", p.APIEndpoint(), strings.TrimLeft(q, "/"))
		go func(wg *sync.WaitGroup, targetURI string) {
			defer wg.Done()

			r := gorequest.New()
			switch strings.ToUpper(method) {
			case "PUT":
				r = r.Put(targetURI)
			case "POST":
				r = r.Post(targetURI)
			case "GET":
				r = r.Get(targetURI)
			}

			resp, _, errs := r.Set("User-Agent", fmt.Sprintf("dbus-%s", dbus.Revision)).End()
			if len(errs) > 0 {
				err = errs[0]
			} else if resp.StatusCode != http.StatusOK {
				err = fmt.Errorf("%s %s", p, http.StatusText(resp.StatusCode))
			}

		}(&wg, targetURI)
	}
	wg.Wait()

	return
}

func (c *controller) LiveParticipants() ([]cluster.Participant, error) {
	participants, marshalled, err := c.zc.ChildrenValues(c.kb.participants())
	if err != nil {
		return nil, err
	}

	var r []cluster.Participant
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
