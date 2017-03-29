package zk

import (
	"github.com/funkygao/dbus/pkg/cluster"
	"github.com/funkygao/go-zookeeper/zk"
	log "github.com/funkygao/log4go"
	"github.com/funkygao/zkclient"
)

var (
	_ cluster.Controller       = &controller{}
	_ zkclient.ZkStateListener = &controller{}
)

type controller struct {
	kb *keyBuilder
	zc *zkclient.Client

	participantID string // in the form of host:port
	weight        int

	leaderID string

	hc      *healthCheck
	elector *leaderElector

	pcl zkclient.ZkChildListener // leader watches alive participants
	rcl zkclient.ZkChildListener // leader watches resources

	// only when participant is leader will this callback be triggered.
	onRebalance func(decision map[string][]string)
}

// New creates a Controller with zookeeper as underlying storage.
func New(zkSvr string, participantID string, weight int, onRebalance func(decision map[string][]string)) cluster.Controller {
	if onRebalance == nil {
		panic("onRebalance nil not allowed")
	}
	if len(zkSvr) == 0 {
		panic("invalid zkSvr")
	}
	if err := validateParticipantID(participantID); err != nil {
		panic(err)
	}

	return &controller{
		kb:            newKeyBuilder(),
		participantID: participantID,
		weight:        weight,
		onRebalance:   onRebalance,
		zc:            zkclient.New(zkSvr, zkclient.WithWrapErrorWithPath()),
	}
}

// NewStandalone creates a controller that will not participate in the cluster election.
// Used for resources definiation.
func NewStandalone(zkSvr string) cluster.Controller {
	c := &controller{
		zc: zkclient.New(zkSvr, zkclient.WithWrapErrorWithPath()),
	}
	c.connectToZookeeper()
	return c
}

func (c *controller) connectToZookeeper() (err error) {
	log.Debug("connecting to zookeeper...")
	if err = c.zc.Connect(); err != nil {
		return
	}

	for retries := 0; retries < 3; retries++ {
		if err = c.zc.WaitUntilConnected(c.zc.SessionTimeout()); err == nil {
			log.Trace("connected to zookeeper")
			break
		}

		log.Warn("retry=%d %v", retries, err)
	}

	return
}

func (c *controller) RegisterResource(input string, resource string) error {
	return c.zc.CreatePersistent(c.kb.resource(resource), []byte(input))
}

func (c *controller) RegisteredResources() (map[string][]string, error) {
	resources, inputs, err := c.zc.ChildrenValues(c.kb.resources())
	if err != nil {
		return nil, err
	}

	r := make(map[string][]string)
	for i, encodedResource := range resources {
		res, _ := c.kb.decodeResource(encodedResource)
		input := string(inputs[i])
		if _, present := r[input]; !present {
			r[input] = []string{res}
		} else {
			r[input] = append(r[input], res)
		}
	}

	return r, nil
}

func (c *controller) Start() (err error) {
	c.pcl = newParticipantChangeListener(c)
	c.rcl = newResourceChangeListener(c)

	if err = c.connectToZookeeper(); err != nil {
		return
	}

	for _, path := range c.kb.persistentKeys() {
		if err = c.zc.CreateEmptyPersistentIfNotPresent(path); err != nil {
			return
		}
	}

	c.hc = newHealthCheck(c.participantID, c.zc, c.kb)
	c.hc.startup()

	c.zc.SubscribeStateChanges(c)

	c.elector = newLeaderElector(c, c.onBecomingLeader, c.onResigningAsLeader)
	c.elector.startup()

	return
}

func (c *controller) Close() (err error) {
	// will delete all ephemeral znodes:
	// participant, controller if leader
	c.zc.Disconnect()

	c.elector.close()
	c.hc.close()

	log.Trace("[%s] controller stopped", c.participantID)
	return
}

func (c *controller) HandleNewSession() (err error) {
	log.Trace("[%s] ZK expired; shutdown all controller components and try re-elect", c.participantID)
	c.onResigningAsLeader()
	c.elector.elect()
	return
}

func (c *controller) HandleStateChanged(state zk.State) (err error) {
	return
}
