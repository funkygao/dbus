package myslave

import (
	"regexp"
)

var (
	expAlterTable = regexp.MustCompile("(?i)^ALTER\\sTABLE\\s.*?`{0,1}(.*?)`{0,1}\\.{0,1}`{0,1}([^`\\.]+?)`{0,1}\\s.*")
)

func isAlterTableQuery(q []byte) (db, table string, yes bool) {
	// FIXME some OnlineSchemaChange tool will not apply 'ALTER TABLE' on the table, but 'RENAME'
	if tuples := expAlterTable.FindSubmatch(q); tuples != nil {
		db = string(tuples[1])
		table = string(tuples[2])
		yes = true
	}

	return
}
