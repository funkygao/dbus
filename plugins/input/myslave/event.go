package myslave

import (
	"fmt"

	"github.com/siddontang/go-mysql/replication"
)

type RowsEvent struct {
	Name     string
	Position uint32

	Schema, Table string
	Action        string
	Rows          [][]interface{}
}

func (r *RowsEvent) String() string {
	return fmt.Sprintf("%s %d %s %s/%s %+v", r.Name, r.Position, r.Action, r.Schema, r.Table, r.Rows)
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
		fmt.Printf("%s not supported", h.EventType)
		return
	}

	r := &RowsEvent{
		Name:     m.pos.Name,
		Position: h.LogPos,
		Schema:   schema,
		Table:    table,
		Action:   action,
		Rows:     e.Rows,
	}
	fmt.Println(r)
}
