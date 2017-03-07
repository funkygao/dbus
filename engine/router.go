package engine

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/funkygao/golib/gofmt"
	log "github.com/funkygao/log4go"
)

// messageRouter is the router/hub shared among all plugins.
type messageRouter struct {
	hub chan *PipelinePack

	stats routerStats

	removeFilterMatcher chan *matcher
	removeOutputMatcher chan *matcher

	filterMatchers []*matcher
	outputMatchers []*matcher
}

func newMessageRouter() *messageRouter {
	return &messageRouter{
		hub:                 make(chan *PipelinePack, Globals().PluginChanSize),
		stats:               routerStats{},
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
		s = fmt.Sprintf("%s %s:%d", s, m.runner.Name(), len(m.InChan()))
		if len(m.InChan()) == globals.PluginChanSize {
			s = fmt.Sprintf("%s(F)", s)
		}
	}
	for _, m := range r.outputMatchers {
		s = fmt.Sprintf("%s %s:%d", s, m.runner.Name(), len(m.InChan()))
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
		pack       *PipelinePack
		ticker     *time.Ticker
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

	log.Trace("Router started with ticker=%s", globals.WatchdogTick)

	ticker = time.NewTicker(globals.WatchdogTick)
	defer ticker.Stop()

LOOP:
	for ok {
		select {
		case matcher = <-r.removeOutputMatcher:
			r.removeMatcher(matcher, r.outputMatchers)

		case matcher = <-r.removeFilterMatcher:
			r.removeMatcher(matcher, r.filterMatchers)

		case <-ticker.C:
			r.stats.render(int(globals.WatchdogTick.Seconds()))
			r.stats.resetPeriodCounters()

		case pack, ok = <-r.hub:
			if !ok {
				globals.Stopping = true
				break LOOP
			}

			r.stats.update(pack) // comment out this line, throughput 1.52M/s -> 1.65M/s

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

type routerStats struct {
	TotalInputMsgN       int64
	PeriodInputMsgN      int32
	TotalInputBytes      int64
	PeriodInputBytes     int64
	TotalProcessedBytes  int64
	TotalProcessedMsgN   int64 // 16 BilionBillion
	PeriodProcessedMsgN  int32
	PeriodProcessedBytes int64
	TotalMaxMsgBytes     int64
	PeriodMaxMsgBytes    int64
}

func (rs *routerStats) update(pack *PipelinePack) {
	atomic.AddInt64(&rs.TotalProcessedMsgN, 1)
	atomic.AddInt32(&rs.PeriodProcessedMsgN, 1)

	if pack.input {
		atomic.AddInt64(&rs.TotalInputMsgN, 1)
		atomic.AddInt32(&rs.PeriodInputMsgN, 1)
		atomic.AddInt64(&rs.TotalInputBytes, int64(pack.Payload.Length()))
		atomic.AddInt64(&rs.PeriodInputBytes, int64(pack.Payload.Length()))
	}
}

func (rs *routerStats) resetPeriodCounters() {
	rs.PeriodProcessedBytes = int64(0)
	rs.PeriodInputBytes = int64(0)
	rs.PeriodInputMsgN = int32(0)
	rs.PeriodProcessedMsgN = int32(0)
	rs.PeriodMaxMsgBytes = int64(0)
}

func (rs *routerStats) render(elapsed int) {
	log.Trace("Total:%10s %10s speed:%6s/s %10s/s max: %s/%s",
		gofmt.Comma(rs.TotalProcessedMsgN),
		gofmt.ByteSize(rs.TotalProcessedBytes),
		gofmt.Comma(int64(rs.PeriodProcessedMsgN/int32(elapsed))),
		gofmt.ByteSize(rs.PeriodProcessedBytes/int64(elapsed)),
		gofmt.ByteSize(rs.PeriodMaxMsgBytes),
		gofmt.ByteSize(rs.TotalMaxMsgBytes))
	log.Trace("Input:%10s %10s speed:%6s/s %10s/s",
		gofmt.Comma(int64(rs.PeriodInputMsgN)),
		gofmt.ByteSize(rs.PeriodInputBytes),
		gofmt.Comma(int64(rs.PeriodInputMsgN/int32(elapsed))),
		gofmt.ByteSize(rs.PeriodInputBytes/int64(elapsed)))
}
