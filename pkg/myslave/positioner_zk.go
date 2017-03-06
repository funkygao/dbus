package myslave

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/funkygao/gafka/zk"
	"github.com/funkygao/golib/sync2"
	zklib "github.com/samuel/go-zookeeper/zk"
)

var _ positioner = &PositionerZk{}

type PositionerZk struct {
	Name   string `json:"name"`
	File   string `json:"file"`
	Offset uint32 `json:"offset"`
	Owner  string `json:"owner"`

	zkzone        *zk.ZkZone
	masterAddr    string
	birthCry      sync2.AtomicBool
	interval      time.Duration
	lastCommitted time.Time

	posPath string // cache
}

func newPositionerZk(name string, zkzone *zk.ZkZone, masterAddr string, interval time.Duration) *PositionerZk {
	return &PositionerZk{
		masterAddr: masterAddr,
		interval:   interval,
		posPath:    posPath(masterAddr),
		zkzone:     zkzone,
		Owner:      myNode(),
		Name:       name,
	}
}

func (z *PositionerZk) MarkAsProcessed(file string, offset uint32) error {
	z.File = file
	z.Offset = offset
	z.birthCry.Set(true)

	now := time.Now()
	if now.Sub(z.lastCommitted) > z.interval {
		return z.Flush()
	}

	return nil
}

func (z *PositionerZk) Flush() (err error) {
	if !z.birthCry.Get() {
		return
	}

	data, _ := json.Marshal(z)
	if _, err = z.zkzone.Conn().Set(z.posPath, data, -1); err == zklib.ErrNoNode {
		_, err = z.zkzone.Conn().Create(z.posPath, data, 0, zklib.WorldACL(zklib.PermAll))
	}

	z.lastCommitted = time.Now() // FIXME race condition with MarkAsProcessed()
	return
}

func (z *PositionerZk) Committed() (file string, offset uint32, err error) {
	var data []byte
	data, _, err = z.zkzone.Conn().Get(z.posPath)
	if err != nil {
		if err == zklib.ErrNoNode {
			data, _ = json.Marshal(z)
			_, err = z.zkzone.Conn().Create(z.posPath, data, 0, zklib.WorldACL(zklib.PermAll))
		}

		return
	}

	if len(strings.TrimSpace(string(data))) == 0 {
		// empty data, return default position
		return
	}

	if err = json.Unmarshal(data, z); err != nil {
		panic(err)
	}

	file = z.File
	offset = z.Offset
	z.Owner = myNode()
	return
}
