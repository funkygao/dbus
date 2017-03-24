package zk

import (
	"github.com/funkygao/go-zookeeper/zk"
	log "github.com/funkygao/log4go"
)

func (c *controller) HandleNewSession() error {
	log.Trace("handling new session")
	return nil
}

func (c *controller) HandleStateChanged(state zk.State) (err error) {
	log.Trace("zk state: %v", state)
	return
}
