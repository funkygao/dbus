package myslave

import (
	"github.com/siddontang/go-mysql/client"
	"github.com/siddontang/go-mysql/mysql"
)

// BinlogRowImage checks MySQL binlog row image, must be in FULL, MINIMAL, NOBLOB.
func (m *MySlave) BinlogRowImage() (string, error) {
	if m.c.String("flavor", "mysql") != mysql.MySQLFlavor {
		return "", nil
	}

	if res, err := m.Execute(`SHOW GLOBAL VARIABLES LIKE "binlog_row_image"`); err != nil {
		return "", err
	} else {
		// MySQL has binlog row image from 5.6, so older will return empty
		return res.GetString(0, 1)
	}
}

// AssertValidRowFormat asserts the mysql master binlog format is ROW.
func (m *MySlave) AssertValidRowFormat() error {
	res, err := m.Execute(`SHOW GLOBAL VARIABLES LIKE "binlog_format";`)
	if err != nil {
		return err
	}

	if f, err := res.GetString(0, 1); err != nil {
		return err
	} else if f != "ROW" {
		return ErrInvalidRowFormat
	}

	return nil
}

// Execute executes a SQL against the mysql master.
func (m *MySlave) Execute(cmd string, args ...interface{}) (rr *mysql.Result, err error) {
	const maxRetry = 3
	for i := 0; i < maxRetry; i++ {
		if m.conn == nil {
			m.conn, err = client.Connect(m.masterAddr, m.user, m.passwd, "")
			if err != nil {
				return nil, err
			}
		}

		rr, err = m.conn.Execute(cmd, args...)
		if err != nil && !mysql.ErrorEqual(err, mysql.ErrBadConn) {
			return
		} else if mysql.ErrorEqual(err, mysql.ErrBadConn) {
			m.conn.Close()
			m.conn = nil
			continue
		} else {
			return
		}
	}

	return
}
