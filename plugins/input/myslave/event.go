package myslave

import (
	"encoding/json"
	"fmt"

	log "github.com/funkygao/log4go"
	"github.com/siddontang/go-mysql/replication"
)

type RowsEvent struct {
	Name      string `json:"log"`
	Position  uint32 `json:"pos"`
	Schema    string `json:"db"`
	Table     string `json:"table"`
	Action    string `json:"action"`
	Timestamp uint32 `json:"t"`

	// binlog has three update event version, v0, v1 and v2.
	// for v1 and v2, the rows number must be even.
	// Two rows for one event, format is [before update row, after update row]
	// for update v0, only one row for a event, and we don't support this version.
	Rows [][]interface{} `json:"rows"`
}

func (r *RowsEvent) String() string {
	return fmt.Sprintf("%s %d %d %s %s/%s %+v", r.Name, r.Position, r.Timestamp, r.Action, r.Schema, r.Table, r.Rows)
}

func (r *RowsEvent) Bytes() []byte {
	b, _ := json.Marshal(r)
	return b
}

func (m *MySlave) handleRowsEvent(h *replication.EventHeader, e *replication.RowsEvent) {
	schema := string(e.Table.Schema)
	table := string(e.Table.Table)

	var action string
	switch h.EventType {
	case replication.WRITE_ROWS_EVENTv1, replication.WRITE_ROWS_EVENTv2:
		action = "insert"

	case replication.DELETE_ROWS_EVENTv1, replication.DELETE_ROWS_EVENTv2:
		action = "delete"

	case replication.UPDATE_ROWS_EVENTv1, replication.UPDATE_ROWS_EVENTv2:
		action = "update"

	default:
		log.Warn("%s not supported: %+v", h.EventType, e)
		return
	}

	m.rowsEvent <- &RowsEvent{
		Name:      m.pos.Name,
		Position:  h.LogPos,
		Schema:    schema,
		Table:     table,
		Action:    action,
		Timestamp: h.Timestamp,
		Rows:      e.Rows,
	}
}
