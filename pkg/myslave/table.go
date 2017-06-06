package myslave

import (
	log "github.com/funkygao/log4go"
)

func (m *MySlave) getTableColumns(db, table string) []string {
	key := db + "." + table
	m.tablesLock.RLock()
	cols, present := m.tables[key]
	m.tablesLock.RUnlock()

	if present {
		return cols
	}

	m.tablesLock.Lock()
	defer m.tablesLock.Unlock()

	if cols, present = m.tables[key]; present {
		return cols
	}

	var err error
	cols, err = m.TableColumns(db, table)
	if err != nil {
		log.Critical("%s.%s get columns: %v", db, table, err)
		return nil
	}

	m.tables[key] = cols
	return cols
}

func (m *MySlave) clearTableCache(db, table string) {
	key := db + "." + table
	m.tablesLock.Lock()
	delete(m.tables, key)
	m.tablesLock.Unlock()
}
