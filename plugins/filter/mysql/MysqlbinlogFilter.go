package mysql

import (
	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/pkg/model"
	conf "github.com/funkygao/jsconf"
	log "github.com/funkygao/log4go"
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
	for pack := range r.InChan() {
		row, ok := pack.Payload.(*model.RowsEvent)
		if !ok {
			pack.Recycle()

			log.Warn("illegal payload: %+v", pack.Payload)
			continue
		}

		p := h.ClonePacket(pack)
		p.Ident = row.Schema
		r.Inject(p)

		pack.Recycle()
	}

	return nil
}
