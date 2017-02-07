package myslave

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/siddontang/go-mysql/replication"
)

func (m *MySlave) Start() error {
	if m.gtid {
		// TODO
	}

	var err error
	if m.pos, err = loadCheckpoint("master.info"); err != nil {
		return err
	}
	if len(m.pos.Addr) != 0 && m.masterAddr != m.pos.Addr {
		// master changed, reset
		m.pos = &checkpoint{Addr: m.masterAddr}
	}

	s, err := m.r.StartSync(m.pos.Pos())
	if err != nil {
		return err
	}

	timeout := time.Second
	for {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		ev, err := s.GetEvent(ctx)
		cancel()

		if err != nil {
			if err == context.DeadlineExceeded {
				if timeout < time.Minute {
					// backoff
					timeout *= 2
				}

				continue
			}

			return err
		}

		// next binglog pos
		//pos = ev.Header.LogPos

		fmt.Printf("==> %T\n", ev.Event)
		switch e := ev.Event.(type) {
		case *replication.RotateEvent:
			//e.NextLogName
			//e.Position
			e.Dump(os.Stdout)

		case *replication.RowsEvent:
			m.handleRowsEvent(ev.Header, e)

		case *replication.QueryEvent:
			fmt.Printf("query: %s\n", string(e.Query))

		case *replication.XIDEvent:
			fmt.Printf("xid: %d\n", e.XID)

		case *replication.FormatDescriptionEvent:
			e.Dump(os.Stdout)

		case *replication.TableMapEvent:
			e.Dump(os.Stdout)

		case *replication.GenericEvent:
			e.Dump(os.Stdout)

		default:
			fmt.Println("unexpected event!!!")
			e.Dump(os.Stdout)
		}
	}

	return nil
}

func (m *MySlave) Close() {
	m.r.Close()
}
