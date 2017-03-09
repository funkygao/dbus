package engine

import (
	"fmt"
)

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
	if pack.Ident == "" {
		panic(fmt.Sprintf("Packet with empty ident: %+v", *pack))
	}

	if len(this.matches) == 0 {
		// match all
		return true
	}

	return this.matches[pack.Ident]
}
