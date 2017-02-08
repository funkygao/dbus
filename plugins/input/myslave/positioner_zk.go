package myslave

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gafka/zk"
	zklib "github.com/samuel/go-zookeeper/zk"
)

var _ positioner = &positionerZk{}

type positionerZk struct {
	File   string `json:"file"`
	Offset uint32 `json:"offset"`

	zkzone     *zk.ZkZone
	masterAddr string

	interval      time.Duration
	lastCommitted time.Time
}

func newPositionerZk(zone string, masterAddr string, interval time.Duration) *positionerZk {
	if len(zone) == 0 {
		panic("zone is required")
	}

	return &positionerZk{
		masterAddr: masterAddr,
		interval:   interval,
		zkzone:     zk.NewZkZone(zk.DefaultConfig(zone, ctx.ZoneZkAddrs(zone))),
	}
}

func (z *positionerZk) MarkAsProcessed(file string, offset uint32) error {
	z.File = file
	z.Offset = offset

	now := time.Now()
	if now.Sub(z.lastCommitted) > z.interval {
		// real commit
		data, _ := json.Marshal(z)
		var err error
		if _, err = z.zkzone.Conn().Set(z.path(), data, -1); err == zklib.ErrNoNode {
			_, err = z.zkzone.Conn().Create(z.path(), data, 0, zklib.WorldACL(zklib.PermAll))
		}

		z.lastCommitted = now
		return err
	}

	return nil
}

func (z *positionerZk) Flush() (err error) {
	data, _ := json.Marshal(z)
	if _, err = z.zkzone.Conn().Set(z.path(), data, -1); err == zklib.ErrNoNode {
		_, err = z.zkzone.Conn().Create(z.path(), data, 0, zklib.WorldACL(zklib.PermAll))
	}

	z.lastCommitted = time.Now()
	return
}

func (z *positionerZk) Committed() (file string, offset uint32, err error) {
	var data []byte
	data, _, err = z.zkzone.Conn().Get(z.path())
	if err != nil {
		if err == zklib.ErrNoNode {
			data, _ = json.Marshal(z)
			_, err = z.zkzone.Conn().Create(z.path(), data, 0, zklib.WorldACL(zklib.PermAll))
		}

		return
	}

	err = json.Unmarshal(data, z)
	file = z.File
	offset = z.Offset
	return
}

func (z *positionerZk) path() string {
	return fmt.Sprintf("/dbus/pos/%s", z.masterAddr)
}
