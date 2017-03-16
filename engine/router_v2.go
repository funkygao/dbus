// +build v2

package engine

import (
	"fmt"
	"sync"
	"time"

	log "github.com/funkygao/log4go"
)

const maxMatchedSubscriber = 100 // TODO validation

// Router is the router/hub shared among all plugins.
type Router struct {
	hub     chan *Packet
	stopper chan struct{}
	m       Matcher
	metrics *routerMetrics
}

func newMessageRouter() *Router {
	return &Router{
		hub:     make(chan *Packet, Globals().HubChanSize),
		stopper: make(chan struct{}),
		metrics: newMetrics(),
		m:       newNaiveMatcher(),
	}
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

	wg.Add(1)
	go r.runReporter(wg)

	var (
		globals = Globals()
		ok      = true
		pack    *Packet
		matcher *matcher
	)

	log.Info("Router started with hub pool=%d", cap(r.hub))

	subs := make([]Subscriber, maxMatchedSubscriber)

LOOP:
	for ok {
		select {
		case pack, ok = <-r.hub:
			if !ok {
				globals.Stopping = true
				// TODO Unsubscribe each
				break LOOP
			}

			if globals.RouterTrack {
				r.metrics.Update(pack) // dryrun throughput 1.8M/s -> 1.3M/s
			}

			// dispatch pack to each subscriber
			subs = subs[0:]
			n := r.m.Lookup(pack.Ident, subs)
			for _, fo := range subs[0:n] {
				fo.InChan() <- pack.incRef()
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
