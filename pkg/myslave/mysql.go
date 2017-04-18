package myslave

import (
	"fmt"

	"github.com/siddontang/go-mysql/client"
	"github.com/siddontang/go-mysql/mysql"
)

// BinlogRowImage checks MySQL binlog row image, must be in FULL, MINIMAL, NOBLOB.
func (m *MySlave) BinlogRowImage() (string, error) {
	if m.c.String("flavor", mysql.MySQLFlavor) != mysql.MySQLFlavor {
		return "", nil
	}

	res, err := m.execute(`SHOW GLOBAL VARIABLES LIKE "binlog_row_image"`)
	if err != nil {
		return "", err
	}

	// MySQL has binlog row image from 5.6, so older will return empty
	return res.GetString(0, 1)
}

// AssertValidRowFormat asserts the mysql master binlog format is ROW.
func (m *MySlave) AssertValidRowFormat() error {
	res, err := m.execute(`SHOW GLOBAL VARIABLES LIKE "binlog_format"`)
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

// BinlogByPos fetches a single binlog event by position.
func (m *MySlave) BinlogByPos(file string, pos int) (*mysql.Result, error) {
	res, err := m.execute(fmt.Sprintf(`SHOW BINLOG EVENTS IN '%s' FROM %d LIMIT 1`, file, pos))
	if err != nil {
		return nil, err
	}

	return res, nil
}

// MasterPosition returns the latest mysql master binlog position info.
func (m *MySlave) MasterPosition() (*mysql.Position, error) {
	rr, err := m.execute("SHOW MASTER STATUS")
	if err != nil {
		return nil, err
	}

	name, _ := rr.GetString(0, 0)
	pos, _ := rr.GetInt(0, 1)
	return &mysql.Position{
		Name: name,
		Pos:  uint32(pos),
	}, nil
}

// MasterBinlogs returns all binlog files on master.
func (m *MySlave) MasterBinlogs() ([]string, error) {
	rr, err := m.execute("SHOW BINARY LOGS")
	if err != nil {
		return nil, err
	}

	names := make([]string, rr.RowNumber())
	for i := 0; i < rr.RowNumber(); i++ {
		name, err := rr.GetString(i, 0) // [0] is Log_name, [1] is File_size
		if err != nil {
			return nil, err
		}

		names[i] = name
	}

	return names, nil
}

// execute executes a SQL against the mysql master.
func (m *MySlave) execute(cmd string, args ...interface{}) (rr *mysql.Result, err error) {
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
