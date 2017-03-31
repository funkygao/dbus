package zk

import (
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/funkygao/assert"
	"github.com/funkygao/dbus/pkg/checkpoint"
	"github.com/funkygao/dbus/pkg/checkpoint/state/binlog"
	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gafka/zk"
	"github.com/funkygao/log4go"
)

func init() {
	log.SetOutput(ioutil.Discard)
	log4go.SetLevel(log4go.ERROR)
	ctx.LoadFromHome()
}

func TestRealPath(t *testing.T) {
	s := binlog.New("")
	assert.Equal(t, "/dbus/checkpoint/myslave/12.12.1.2:3334", realPath(s, "12.12.1.2:3334"))
}

func TestCheckpointZKBinlog(t *testing.T) {
	zone := ctx.DefaultZone()
	zkzone := zk.NewZkZone(zk.DefaultConfig(zone, ctx.ZoneZkAddrs(zone)))
	s := binlog.New("")
	zpath := "12.12.1.2:3334"
	z := New(zkzone, s, zpath, time.Minute)
	defer zkzone.DeleteRecursive(realPath(s, zpath))

	assert.Equal(t, checkpoint.ErrStateNotFound, z.LastPersistedState(s))

	s.File = "f1"
	s.Offset = 5
	assert.Equal(t, nil, z.Commit(s))
	assert.Equal(t, nil, z.Shutdown())

	s = binlog.New("")
	assert.Equal(t, nil, z.LastPersistedState(s))
	assert.Equal(t, "f1", s.File)
	assert.Equal(t, uint32(5), s.Offset)
}
