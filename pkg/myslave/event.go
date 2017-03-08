package myslave

import (
	"github.com/funkygao/dbus/pkg/model"
	log "github.com/funkygao/log4go"
	"github.com/siddontang/go-mysql/replication"
)

func (m *MySlave) handleRowsEvent(f string, h *replication.EventHeader, e *replication.RowsEvent) {
	schema := string(e.Table.Schema)
	table := string(e.Table.Table)
	if !m.predicate(schema, table) {
		log.Debug("[%s] ignored[%s.%s]: %+v %+v", m.masterAddr, schema, table, h, e)
		m.p.MarkAsProcessed(f, h.LogPos) // FIXME batcher partial failure?
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
		log.Warn("[%s] %s not supported: %+v", m.masterAddr, h.EventType, e)
		return
	}

	m.rowsEvent <- &model.RowsEvent{
		Log:       f,
		Position:  h.LogPos, // next binlog pos
		Schema:    schema,
		Table:     table,
		Action:    action,
		Timestamp: h.Timestamp,
		Rows:      e.Rows,
	}
}

func (m *MySlave) predicate(schema, table string) bool {
	if len(m.db) > 0 && m.db != schema {
		return false
	}
	if _, present := m.dbExcluded[schema]; present {
		return false
	}
	if _, present := m.tableExcluded[table]; present {
		return false
	}

	return true
}
