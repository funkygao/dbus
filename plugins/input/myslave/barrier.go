package myslave

import (
	"bytes"
	"time"

	"github.com/funkygao/dbus"
	log "github.com/funkygao/log4go"
	"github.com/samuel/go-zookeeper/zk"
)

func (m *MySlave) leaveCluster() {
	if err := m.z.Conn().Delete(myNodePath(m.masterAddr), -1); err != nil {
		log.Error("[%s] %s", m.name, err)
	}

	masterData := []byte(myNode())
	data, stat, err := m.z.Conn().Get(masterPath(m.masterAddr))
	if err != nil {
		log.Error("[%s] %s", m.name, err)
		return
	}

	if bytes.Equal(data, masterData) {
		// I'm the master
		if err := m.z.Conn().Delete(masterPath(m.masterAddr), stat.Version); err != nil {
			log.Error("[%s] %s", m.name, err)
		}
	}
}

// TODO session expire
func (m *MySlave) joinClusterAndBecomeMaster() {
	// become present
	backoff := time.Second
	for {
		if err := m.z.CreateEphemeralZnode(myNodePath(m.masterAddr), []byte(dbus.BuildID)); err != nil {
			log.Error("[%s] unable present: %s", m.name, err)

			time.Sleep(backoff)

			if backoff < time.Minute {
				backoff *= 2
			}
		} else {
			log.Trace("[%s] become present", m.name)
			break
		}
	}

	// become master
	masterData := []byte(myNode())
	for {
		if err := m.z.CreateEphemeralZnode(masterPath(m.masterAddr), masterData); err != nil {
			if err != zk.ErrNodeExists {
				log.Error("[%s] become master: %s", m.name, err)
			}

			time.Sleep(time.Minute)
		} else {
			log.Trace("[%s] become master", m.name)
			return
		}
	}
}
