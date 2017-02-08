package myslave

import (
	"encoding/json"
	"fmt"

	"github.com/funkygao/dbus/engine"
	log "github.com/funkygao/log4go"
	"github.com/siddontang/go-mysql/replication"
)

var _ engine.Payloader = &RowsEvent{}

// RowsEvent is a structured mysql binlog rows event.
// It implements engine.Payloader interface and can be transferred between plugins.
type RowsEvent struct {
	Log       string `json:"log"`
	Position  uint32 `json:"pos"`
	Schema    string `json:"db"`
	Table     string `json:"tbl"`
	Action    string `json:"dml"`
	Timestamp uint32 `json:"ts"`

	// binlog has three update event version, v0, v1 and v2.
	// for v1 and v2, the rows number must be even.
	// Two rows for one event, format is [before update row, after update row]
	// for update v0, only one row for a event, and we don't support this version.
	Rows [][]interface{} `json:"rows"`
}

// Implements Payloader.
func (r *RowsEvent) String() string {
	return fmt.Sprintf("%s %d %d %s %s/%s %+v", r.Log, r.Position, r.Timestamp, r.Action, r.Schema, r.Table, r.Rows)
}

// Implements Payloader.
func (r *RowsEvent) Bytes() []byte {
	b, _ := json.Marshal(r)
	return b
}

// Implements Payloader.
func (r *RowsEvent) Length() int {
	return len(r.Log) + len(r.Schema) + len(r.Table) + 9 // FIXME Rows not counted
}

func (m *MySlave) handleRowsEvent(f string, h *replication.EventHeader, e *replication.RowsEvent) {
	schema := string(e.Table.Schema)
	table := string(e.Table.Table)

	var action string
	switch h.EventType {
	case replication.WRITE_ROWS_EVENTv1, replication.WRITE_ROWS_EVENTv2:
		action = "I"

	case replication.DELETE_ROWS_EVENTv1, replication.DELETE_ROWS_EVENTv2:
		action = "D"

	case replication.UPDATE_ROWS_EVENTv1, replication.UPDATE_ROWS_EVENTv2:
		action = "U"

	default:
		log.Warn("%s not supported: %+v", h.EventType, e)
		return
	}

	m.rowsEvent <- &RowsEvent{
		Log:       f,
		Position:  h.LogPos,
		Schema:    schema,
		Table:     table,
		Action:    action,
		Timestamp: h.Timestamp,
		Rows:      e.Rows,
	}
}
