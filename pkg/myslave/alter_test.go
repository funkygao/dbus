package myslave

import (
	"testing"

	"github.com/funkygao/assert"
)

func TestIsAlterTableQuery(t *testing.T) {
	db, table, yes := isAlterTableQuery([]byte("commit"))
	assert.Equal(t, false, yes)

	db, table, yes = isAlterTableQuery([]byte("alTer table foo add id int"))
	assert.Equal(t, "", db)
	assert.Equal(t, "foo", table)
	assert.Equal(t, true, yes)

	db, table, yes = isAlterTableQuery([]byte("insert into bar(id) values(4); alTer table foo add id int"))
	assert.Equal(t, false, yes)
}

// 130ns/op
func BenchmarkIsAlterTableQueryNo(b *testing.B) {
	q := []byte("commit")
	for i := 0; i < b.N; i++ {
		isAlterTableQuery(q)
	}
}

// 1200ns/op
func BenchmarkIsAlterTableQueryYes(b *testing.B) {
	q := []byte("alTer table foo add id int")
	for i := 0; i < b.N; i++ {
		isAlterTableQuery(q)
	}
}
