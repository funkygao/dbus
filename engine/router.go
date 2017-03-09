package engine

import (
	"fmt"
	"sync"
	"time"

	log "github.com/funkygao/log4go"
)

// messageRouter is the router/hub shared among all plugins.
type messageRouter struct {
	hub chan *Packet

	metrics *routerMetrics

	removeFilterMatcher chan *matcher
	removeOutputMatcher chan *matcher

	filterMatchers []*matcher
	outputMatchers []*matcher
}

func newMessageRouter() *messageRouter {
	return &messageRouter{
		hub:                 make(chan *Packet, Globals().PluginChanSize),
		metrics:             newMetrics(),
		removeFilterMatcher: make(chan *matcher),
		removeOutputMatcher: make(chan *matcher),
		filterMatchers:      make([]*matcher, 0, 10),
		outputMatchers:      make([]*matcher, 0, 10),
	}
}

func (r *messageRouter) addFilterMatcher(matcher *matcher) {
	r.filterMatchers = append(r.filterMatchers, matcher)
}

func (r *messageRouter) addOutputMatcher(matcher *matcher) {
	r.outputMatchers = append(r.outputMatchers, matcher)
}

func (r *messageRouter) reportMatcherQueues() {
	globals := Globals()
	s := fmt.Sprintf("Queued hub=%d", len(r.hub))
	if len(r.hub) == globals.PluginChanSize {
		s = fmt.Sprintf("%s(F)", s)
	}

	for _, m := range r.filterMatchers {
		s = fmt.Sprintf("%s %s=%d", s, m.runner.Name(), len(m.InChan()))
		if len(m.InChan()) == globals.PluginChanSize {
			s = fmt.Sprintf("%s(F)", s)
		}
	}
	for _, m := range r.outputMatchers {
		s = fmt.Sprintf("%s %s=%d", s, m.runner.Name(), len(m.InChan()))
		if len(m.InChan()) == globals.PluginChanSize {
			s = fmt.Sprintf("%s(F)", s)
		}
	}

	log.Trace(s)
}

// Dispatch pack from Input to MatchRunners
func (r *messageRouter) Start(wg *sync.WaitGroup) {
	defer wg.Done()

	var (
		globals    = Globals()
		ok         = true
		pack       *Packet
		matcher    *matcher
		foundMatch bool
	)

	go func() {
		t := time.NewTicker(globals.WatchdogTick)
		defer t.Stop()

		for range t.C {
			r.reportMatcherQueues()
		}
	}()

	log.Info("Router started")

LOOP:
	for ok {
		select {
		case matcher = <-r.removeOutputMatcher:
			r.removeMatcher(matcher, r.outputMatchers)

		case matcher = <-r.removeFilterMatcher:
			r.removeMatcher(matcher, r.filterMatchers)

		case pack, ok = <-r.hub:
			if !ok {
				globals.Stopping = true
				break LOOP
			}

			if globals.RouterTrack {
				r.metrics.Update(pack) // dryrun throughput 1.8M/s -> 1.3M/s
			}

			foundMatch = false

			// If we send pack to filterMatchers and then outputMatchers
			// because filter may change pack Ident, and this pack because
			// of shared mem, may match both filterMatcher and outputMatcher
			// then dup dispatching happens!!!
			//
			// We have to dispatch to Output then Filter to avoid that case
			for _, matcher = range r.outputMatchers {
				// a pack can match several Output
				if matcher != nil && matcher.Match(pack) {
					foundMatch = true

					pack.incRef()
					matcher.InChan() <- pack
				}
			}

			// got pack from Input, now dispatch
			// for each target, pack will inc ref count
			// and the router will dec ref count only once
			for _, matcher = range r.filterMatchers {
				// a pack can match several Filter
				if matcher != nil && matcher.Match(pack) {
					foundMatch = true

					pack.incRef()
					matcher.InChan() <- pack
				}
			}

			if !foundMatch {
				// Maybe we closed all filter/output inChan, but there
				// still exits some remnant packs in router.hub.
				// To handle r issue, Input/Output should be stateful.
				log.Debug("no match: %+v", pack)
			}

			// never forget this!
			// if no sink found, this packet is recycled directly for latter use
			pack.Recycle()
		}
	}

}

func (r *messageRouter) removeMatcher(matcher *matcher, matchers []*matcher) {
	for idx, m := range matchers {
		if m == matcher {
			log.Trace("closing matcher for %s", m.runner.Name())

			close(m.InChan())
			matchers[idx] = nil
			return
		}
	}
}
