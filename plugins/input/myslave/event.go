package myslave

import (
	"github.com/siddontang/go-mysql/replication"
)

func (m *MySlave) handleRowsEvent(h *replication.EventHeader, e *replication.RowsEvent) {

}
