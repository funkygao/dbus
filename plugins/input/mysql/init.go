package mysql

import (
	"github.com/funkygao/dbus/engine"
)

var (
	_ engine.Input = &MysqlbinlogInput{}
)

func init() {
	engine.RegisterPlugin("MysqlbinlogInput", func() engine.Plugin {
		return new(MysqlbinlogInput)
	})
}
