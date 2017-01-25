package engine

import (
	"fmt"
)

type Matcher struct {
	runner  FilterOutputRunner
	matches map[string]bool
}

func NewMatcher(matches []string, r FilterOutputRunner) *Matcher {
	this := new(Matcher)
	this.matches = make(map[string]bool)
	for _, m := range matches {
		this.matches[m] = true
	}
	this.runner = r
	return this
}

func (this *Matcher) InChan() chan *PipelinePack {
	return this.runner.InChan()
}

func (this *Matcher) match(pack *PipelinePack) bool {
	if pack.Ident == "" {
		errmsg := fmt.Sprintf("Pack with empty ident: %s", *pack)
		panic(errmsg)
	}

	if len(this.matches) == 0 {
		// match all
		return true
	}

	return this.matches[pack.Ident]
}
