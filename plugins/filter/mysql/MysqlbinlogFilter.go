package filter

import (
	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/pkg/model"
	conf "github.com/funkygao/jsconf"
	log "github.com/funkygao/log4go"
)

var (
	_ engine.Filter = &MysqlbinlogFilter{}
)

/*

        +-------+
        | mysql |
        +-------+
            | rows event
   +-------------------+
   |       |           |
  db1     db2         dbN
   |       |           |
   +-------------------+
           |
     MysqlbinlogFilter
           | dispatch
   +-------------------+
   |       |           |
 out1     out2       outM
*/
type MysqlbinlogFilter struct {
}

func (this *MysqlbinlogFilter) Init(config *conf.Conf) {}

func (this *MysqlbinlogFilter) Run(r engine.FilterRunner, h engine.PluginHelper) error {
	for {
		select {
		case pack, ok := <-r.InChan():
			if !ok {
				return nil
			}

			row, ok := pack.Payload.(*model.RowsEvent)
			if !ok {
				log.Warn("illegal payload: %+v", pack.Payload)
				continue
			}

			p := h.ClonePacket(pack)
			p.Ident = row.Schema
			r.Inject(p)

			pack.Recycle()
		}
	}
}

func init() {
	engine.RegisterPlugin("MysqlbinlogFilter", func() engine.Plugin {
		return new(MysqlbinlogFilter)
	})
}
