// +build !v2

package engine

// matcher belongs to the singleton router, it requires no lock.
type matcher struct {
	runner  *foRunner
	matches map[string]bool
}

func newMatcher(matches []string, r *foRunner) *matcher {
	m := new(matcher)
	m.matches = make(map[string]bool)
	for _, match := range matches {
		m.matches[match] = true
	}
	m.runner = r
	return m
}

func (m *matcher) InChan() chan *Packet {
	return m.runner.inChan
}

func (m *matcher) Match(pack *Packet) bool {
	return m.matches[pack.Ident]
}
