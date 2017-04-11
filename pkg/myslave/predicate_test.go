package myslave

import (
	"testing"

	"github.com/funkygao/assert"
)

func TestPredicate(t *testing.T) {
	m := New("", "", "")
	m.setupPredicate()
	assert.Equal(t, true, m.Predicate("db1", ""))

	m.dbAllowed = map[string]struct{}{
		"db2": {},
	}
	m.setupPredicate()
	assert.Equal(t, false, m.Predicate("db1", ""))

	delete(m.dbAllowed, "db2") // reset
	m.dbExcluded = map[string]struct{}{
		"db2": {},
	}
	m.setupPredicate()
	assert.Equal(t, true, m.Predicate("db1", ""))
}

func BenchmarkPredicateAllow(b *testing.B) {
	m := New("", "", "")
	m.dbAllowed = map[string]struct{}{
		"db2": {},
	}
	m.setupPredicate()

	for i := 0; i < b.N; i++ {
		m.Predicate("db2", "table")
	}
}

func BenchmarkPredicateDisallow(b *testing.B) {
	m := New("", "", "")
	m.dbAllowed = map[string]struct{}{
		"db1": {},
	}
	m.setupPredicate()
	for i := 0; i < b.N; i++ {
		m.Predicate("db2", "table")
	}
}
