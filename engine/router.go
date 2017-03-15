// +build !v2

package engine

import (
	"fmt"
	"sync"
	"time"

	log "github.com/funkygao/log4go"
)

/*
Router is the router/hub shared among all plugins which dispatches
packet along the plugins.

A normal packet lifecycle:

        +-------------+
   	+-- | | | | | | | | Input(ipool)
    |   +-------------+
    |          |
    |   +-------------+
    |   | | | | | | | | Hub(hpool)
    |   +-------------+
    |          |
    |          |-------------------------------------+
    |          |                                     |
    |   +-------------+                       +-------------+
    |   | | | | | | | | Output/Filter(ppool)  | | | | | | | | Output/Filter(ppool)
    |   +-------------+                       +-------------+
    |          |                                     |
    |          +-------------------------------------+
    |                        |
	+-------<----------------+
	     Recycle


A normal cloned packet lifecycle:

        +-------------+
   	+-- | | | | | | | | Engine(fpool)
    |   +-------------+
    |          |
    |          | ClonePacket
    |          |
    |   +-------------+
    |   | | | | | | | | Filter
    |   +-------------+
    |          |
    |   +-------------+
    |   | | | | | | | | Output/Filter(ppool)
    |   +-------------+
    |          |
	+-------<--+
	     Recycle


*/
type Router struct {
	hub     chan *Packet
	stopper chan struct{}

	metrics *routerMetrics

	removeFilterMatcher chan *matcher
	removeOutputMatcher chan *matcher

	filterMatchers []*matcher
	outputMatchers []*matcher
}

func newMessageRouter() *Router {
	return &Router{
		hub:                 make(chan *Packet, Globals().HubChanSize),
		stopper:             make(chan struct{}),
		metrics:             newMetrics(),
		removeFilterMatcher: make(chan *matcher),
		removeOutputMatcher: make(chan *matcher),
		filterMatchers:      make([]*matcher, 0, 10),
		outputMatchers:      make([]*matcher, 0, 10),
	}
}

func (r *Router) addFilterMatcher(matcher *matcher) {
	r.filterMatchers = append(r.filterMatchers, matcher)
}

func (r *Router) addOutputMatcher(matcher *matcher) {
	r.outputMatchers = append(r.outputMatchers, matcher)
}

func (r *Router) reportMatcherQueues() {
	globals := Globals()
	s := fmt.Sprintf("Queued hub=%d", len(r.hub))
	if len(r.hub) == globals.HubChanSize {
		s = fmt.Sprintf("%s(F)", s)
	}

	for _, fm := range r.filterMatchers {
		s = fmt.Sprintf("%s %s=%d", s, fm.runner.Name(), len(fm.InChan()))
		if len(fm.InChan()) == globals.PluginChanSize {
			s = fmt.Sprintf("%s(F)", s)
		}
	}
	for _, om := range r.outputMatchers {
		s = fmt.Sprintf("%s %s=%d", s, om.runner.Name(), len(om.InChan()))
		if len(om.InChan()) == globals.PluginChanSize {
			s = fmt.Sprintf("%s(F)", s)
		}
	}

	log.Trace(s)
}

// Dispatch pack from Input to MatchRunners
func (r *Router) Start(wg *sync.WaitGroup) {
	defer wg.Done()

	var (
		globals    = Globals()
		ok         = true
		pack       *Packet
		matcher    *matcher
		foundMatch bool
	)

	wg.Add(1)
	go r.runReporter(wg)

	log.Info("Router started with hub pool=%d", cap(r.hub))

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

			// dispatch pack to output plugins, 1 to many
			for _, matcher = range r.outputMatchers {
				if matcher != nil && matcher.Match(pack) {
					foundMatch = true

					matcher.InChan() <- pack.incRef()
				}
			}

			// dispatch pack to filter plugins, 1 to many
			for _, matcher = range r.filterMatchers {
				if matcher != nil && matcher.Match(pack) {
					foundMatch = true

					matcher.InChan() <- pack.incRef()
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

func (r *Router) Stop() {
	log.Trace("Router stopping...")
	close(r.hub)
	close(r.stopper)

	for ident, m := range r.metrics.m {
		log.Trace("routed to [%s] %d", ident, m.Count())
	}
}

func (r *Router) runReporter(wg *sync.WaitGroup) {
	defer wg.Done()

	t := time.NewTicker(Globals().WatchdogTick)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			r.reportMatcherQueues()

		case <-r.stopper:
			return
		}
	}

}

func (r *Router) removeMatcher(matcher *matcher, matchers []*matcher) {
	for idx, m := range matchers {
		if m == matcher {
			log.Trace("closing matcher for %s", m.runner.Name())

			close(m.InChan())
			matchers[idx] = nil
			return
		}
	}
}
