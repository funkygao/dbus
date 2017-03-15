// +build !v2

package engine

// matcher belongs to the singleton router, it requires no lock.
type matcher struct {
	runner  FilterOutputRunner
	matches map[string]bool
}

func newMatcher(matches []string, r FilterOutputRunner) *matcher {
	this := new(matcher)
	this.matches = make(map[string]bool)
	for _, m := range matches {
		this.matches[m] = true
	}
	this.runner = r
	return this
}

func (this *matcher) InChan() chan *Packet {
	return this.runner.InChan()
}

func (this *matcher) Match(pack *Packet) bool {
	return this.matches[pack.Ident]
}
