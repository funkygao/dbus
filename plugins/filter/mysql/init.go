package mysql

import (
	"github.com/funkygao/dbus/engine"
)

var (
	_ engine.Filter = &MysqlbinlogFilter{}
)

func init() {
	engine.RegisterPlugin("MysqlbinlogFilter", func() engine.Plugin {
		return new(MysqlbinlogFilter)
	})
}
