package zk

import (
	log "github.com/funkygao/log4go"
)

func (c *controller) connectToZookeeper(zkSvr string) error {
	log.Debug("connecting to %s...", zkSvr)
	c.zc.SubscribeStateChanges(c)

	if err := c.zc.Connect(); err != nil {
		return err
	}

	var err error
	for retries := 0; retries < 3; retries++ {
		if err = c.zc.WaitUntilConnected(c.zc.SessionTimeout()); err == nil {
			break
		}

		log.Warn("%s retry=%d %v", zkSvr, retries, err)
	}

	log.Debug("connected to %s", zkSvr)
	return nil
}
