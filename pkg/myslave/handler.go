package myslave

import (
	"time"

	"github.com/funkygao/dbus/pkg/model"
	log "github.com/funkygao/log4go"
	"github.com/siddontang/go-mysql/replication"
)

func (m *MySlave) handleRowsEvent(f string, h *replication.EventHeader, e *replication.RowsEvent) {
	schema := string(e.Table.Schema)
	table := string(e.Table.Table)
	if !m.Predicate(schema, table) {
		log.Debug("[%s] ignored[%s.%s]: %+v %+v", m.dsn, schema, table, h, e)
		m.CommitPosition(f, h.LogPos) // FIXME batcher partial failure?
		return
	}

	var action string
	switch h.EventType {
	case replication.WRITE_ROWS_EVENTv1, replication.WRITE_ROWS_EVENTv2:
		action = "I"

	case replication.DELETE_ROWS_EVENTv1, replication.DELETE_ROWS_EVENTv2:
		action = "D"

	case replication.UPDATE_ROWS_EVENTv1, replication.UPDATE_ROWS_EVENTv2:
		action = "U"

	default:
		log.Warn("[%s] %s not supported: %+v", m.dsn, h.EventType, e)
		return
	}

	rowsEvent := &model.RowsEvent{
		Log:           f,
		Position:      h.LogPos, // next binlog pos
		Schema:        schema,
		Table:         table,
		Action:        action,
		Timestamp:     h.Timestamp,
		DbusTimestamp: time.Now().UnixNano(),
		Columns:       m.getTableColumns(schema, table),
		Rows:          e.Rows,
	}
	m.rowsEvent <- rowsEvent.SetFlags(e.Flags)
}
