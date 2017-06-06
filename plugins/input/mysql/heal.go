package mysql

import (
	"strings"

	"github.com/funkygao/dbus/pkg/myslave"
	log "github.com/funkygao/log4go"
)

func (this *MysqlbinlogInput) tryAutoHeal(name string, err error, slave *myslave.MySlave) {
	if strings.Contains(err.Error(), "ERROR 1236 (HY000)") {
		log.Trace("[%s] auto healing ERROR 1236", name)
		if pos, _err := slave.MasterPosition(); _err == nil {
			// FIXME the pos might miss 'table id' info.
			log.Warn("[%s] reset %s position to: %+v", name, slave.DSN(), pos)
			if er := slave.CommitPosition(pos.Name, 4); er != nil {
				log.Error("[%s] %s: %v", name, slave.DSN(), er)
			}
		} else {
			log.Error("[%s] %s", name, _err)
		}
	}
}
