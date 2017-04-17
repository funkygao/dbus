package zk

import (
	"path"
	"time"

	"github.com/funkygao/dbus/pkg/checkpoint"
	"github.com/funkygao/gafka/zk"
	"github.com/funkygao/golib/sync2"
	zklib "github.com/samuel/go-zookeeper/zk"
)

var (
	_ checkpoint.Checkpoint = &checkpointZK{}
)

type checkpointZK struct {
	path     string
	zkzone   *zk.ZkZone
	interval time.Duration

	birthCry      sync2.AtomicBool
	lastState     checkpoint.State
	lastCommitted time.Time
}

func New(zkzone *zk.ZkZone, state checkpoint.State, zroot string, zpath string, interval time.Duration) checkpoint.Checkpoint {
	if len(zroot) > 0 {
		root = zroot
	}
	zpath = realPath(state, zpath)
	if err := zkzone.EnsurePathExists(path.Dir(zpath)); err != nil {
		panic(err)
	}

	return &checkpointZK{
		interval: interval,
		zkzone:   zkzone,
		path:     zpath,
	}
}

func (z *checkpointZK) Shutdown() error {
	return z.persist()
}

func (z *checkpointZK) Commit(state checkpoint.State) error {
	z.lastState = state // TODO what if rewind?
	z.birthCry.Set(true)

	now := time.Now()
	if now.Sub(z.lastCommitted) > z.interval {
		return z.persist()
	}

	return nil
}

func (z *checkpointZK) LastPersistedState(state checkpoint.State) (err error) {
	var data []byte
	data, _, err = z.zkzone.Conn().Get(z.path)
	if err != nil {
		if err == zklib.ErrNoNode {
			err = checkpoint.ErrStateNotFound
		}

		return
	}

	state.Unmarshal(data)
	return
}

func (z *checkpointZK) persist() (err error) {
	if !z.birthCry.Get() {
		return
	}

	data := z.lastState.Marshal()
	if _, err = z.zkzone.Conn().Set(z.path, data, -1); err == zklib.ErrNoNode {
		_, err = z.zkzone.Conn().Create(z.path, data, 0, zklib.WorldACL(zklib.PermAll))
	}

	if err != nil {
		// TODO resort to local disk
	}

	z.lastCommitted = time.Now()
	return
}
