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

    +--------->---------------+
    |                         |
    |                         |
    |          +-------------------------------+
    |          |                               |
    |   +-------------+                 +-------------+
   	|   | | | | | | | | Input(ipool)    | | | | | | | | Input(ipool)
    |   +-------------+                 +-------------+
    |          |                               |
    |          +-------------------------------+
    |                          |
    |                   +-------------+
    |                   | | | | | | | | Hub(hpool) <------------------------------------+
    |                   +-------------+													|
    |                          |														|
    |          |-------------------------------------+									|
    |          |                                     |									|
    |   +-------------+                       +-------------+							|
    |   | | | | | | | | Output/Filter(ppool)  | | | | | | | | Output/Filter(ppool)		|
    |   +-------------+                       +-------------+							|
    |          |                                     |									|
    |          +-------------------------------------+									|
    |                        |    |														|
	+-------<----------------+	  +------------->---------------------------------------+
	     Recycle                             Inject


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
	stopper chan struct{}
	metrics *routerMetrics

	hub chan *Packet

	filterMatchers []*matcher
	outputMatchers []*matcher
}

func newRouter() *Router {
	return &Router{
		hub:            make(chan *Packet, Globals().HubChanSize),
		stopper:        make(chan struct{}),
		metrics:        newMetrics(),
		filterMatchers: make([]*matcher, 0, 10),
		outputMatchers: make([]*matcher, 0, 10),
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
	full := false
	s := fmt.Sprintf("Queued hub=%d/%d", len(r.hub), globals.HubChanSize)
	if len(r.hub) == globals.HubChanSize {
		s = fmt.Sprintf("%s(F)", s)
		full = true
	}

	for _, fm := range r.filterMatchers {
		s = fmt.Sprintf("%s %s=%d/%d", s, fm.runner.Name(), len(fm.InChan()), globals.FilterRecyclePoolSize)
		if len(fm.InChan()) == globals.PluginChanSize {
			s = fmt.Sprintf("%s(F)", s)
			full = true
		}
	}
	for _, om := range r.outputMatchers {
		s = fmt.Sprintf("%s %s=%d/%d", s, om.runner.Name(), len(om.InChan()), globals.PluginChanSize)
		if len(om.InChan()) == globals.PluginChanSize {
			s = fmt.Sprintf("%s(F)", s)
			full = true
		}
	}

	if full {
		log.Trace(s)
	}
}

// Start starts the router: dispatch pack from Input to MatchRunners.
func (r *Router) Start(wg *sync.WaitGroup) {
	defer wg.Done()

	wg.Add(1)
	go r.runReporter(wg)

	log.Info("Router started with hub pool=%d", cap(r.hub))

	var (
		globals = Globals()
		pack    *Packet
	)
	for {
		select {
		case pack = <-r.hub:
			// the packet can be from: Input|Filter|Output
			if globals.RouterTrack {
				r.metrics.Update(pack) // dryrun throughput 2.1M/s -> 1.6M/s
			}

			foundMatch := false
			// dispatch pack to output plugins, 1 to many
			for _, matcher := range r.outputMatchers {
				if matcher != nil && matcher.Match(pack) {
					foundMatch = true

					matcher.InChan() <- pack.incRef()
				}
			}

			// dispatch pack to filter plugins, 1 to many
			for _, matcher := range r.filterMatchers {
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

		case <-r.stopper:
			// now Input has all stopped
			// start to drain in-flight packets from Filter|Output plugins
			for {
				select {
				case pack = <-r.hub:
					if globals.RouterTrack {
						r.metrics.Update(pack)
					}

					foundMatch := false
					for _, matcher := range r.outputMatchers {
						if matcher != nil && matcher.Match(pack) {
							foundMatch = true

							matcher.InChan() <- pack.incRef()
						}
					}
					for _, matcher := range r.filterMatchers {
						if matcher != nil && matcher.Match(pack) {
							foundMatch = true

							matcher.InChan() <- pack.incRef()
						}
					}
					if !foundMatch {
						log.Debug("no match: %+v", pack)
					}

					pack.Recycle()

				default:
					log.Trace("Router fully drained, stopping Filter|Output plugins...")

					// async notify Filter|Output plugins to stop
					for _, fm := range r.filterMatchers {
						close(fm.InChan())
					}
					r.filterMatchers = nil
					for _, om := range r.outputMatchers {
						close(om.InChan())
					}
					r.outputMatchers = nil
					close(r.hub)
					return
				}
			}

		}
	}

}

func (r *Router) Stop() {
	log.Debug("Router stopping...")
	close(r.stopper)

	if Globals().RouterTrack {
		for ident, m := range r.metrics.m {
			log.Trace("routed to [%s] %d", ident, m.Count())
		}
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
			r.reportMatcherQueues()
			return
		}
	}
}
