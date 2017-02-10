package engine

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/funkygao/golib/gofmt"
	log "github.com/funkygao/log4go"
)

type messageRouter struct {
	hub chan *PipelinePack

	stats routerStats

	removeFilterMatcher chan *matcher
	removeOutputMatcher chan *matcher

	filterMatchers []*matcher
	outputMatchers []*matcher
}

func newMessageRouter() (this *messageRouter) {
	this = new(messageRouter)
	this.hub = make(chan *PipelinePack, Globals().PluginChanSize)
	this.stats = routerStats{}
	this.removeFilterMatcher = make(chan *matcher)
	this.removeOutputMatcher = make(chan *matcher)
	this.filterMatchers = make([]*matcher, 0, 10)
	this.outputMatchers = make([]*matcher, 0, 10)

	return
}

func (this *messageRouter) addFilterMatcher(matcher *matcher) {
	this.filterMatchers = append(this.filterMatchers, matcher)
}

func (this *messageRouter) addOutputMatcher(matcher *matcher) {
	this.outputMatchers = append(this.outputMatchers, matcher)
}

func (this *messageRouter) reportMatcherQueues() {
	globals := Globals()
	s := fmt.Sprintf("Queued hub=%d", len(this.hub))
	if len(this.hub) == globals.PluginChanSize {
		s = fmt.Sprintf("%s(F)", s)
	}

	for _, m := range this.filterMatchers {
		s = fmt.Sprintf("%s %s:%d", s, m.runner.Name(), len(m.InChan()))
		if len(m.InChan()) == globals.PluginChanSize {
			s = fmt.Sprintf("%s(F)", s)
		}
	}
	for _, m := range this.outputMatchers {
		s = fmt.Sprintf("%s %s:%d", s, m.runner.Name(), len(m.InChan()))
		if len(m.InChan()) == globals.PluginChanSize {
			s = fmt.Sprintf("%s(F)", s)
		}
	}

	log.Trace(s)
}

// Dispatch pack from Input to MatchRunners
func (this *messageRouter) Start() {
	var (
		globals    = Globals()
		ok         = true
		pack       *PipelinePack
		ticker     *time.Ticker
		matcher    *matcher
		foundMatch bool
	)

	ticker = time.NewTicker(globals.WatchdogTick)
	defer ticker.Stop()

	go func() {
		t := time.NewTicker(globals.WatchdogTick)
		defer t.Stop()

		for range t.C {
			this.reportMatcherQueues()
		}
	}()

	log.Trace("Router started with ticker=%s", globals.WatchdogTick)

LOOP:
	for ok {
		select {
		case matcher = <-this.removeOutputMatcher:
			this.removeMatcher(matcher, this.outputMatchers)

		case matcher = <-this.removeFilterMatcher:
			this.removeMatcher(matcher, this.filterMatchers)

		case <-ticker.C:
			this.stats.render(int(globals.WatchdogTick.Seconds()))
			this.stats.resetPeriodCounters()

		case pack, ok = <-this.hub:
			if !ok {
				globals.Stopping = true
				break LOOP
			}

			this.stats.update(pack)

			pack.diagnostics.Reset()
			foundMatch = false

			// If we send pack to filterMatchers and then outputMatchers
			// because filter may change pack Ident, and this pack because
			// of shared mem, may match both filterMatcher and outputMatcher
			// then dup dispatching happens!!!
			//
			// We have to dispatch to Output then Filter to avoid that case
			for _, matcher = range this.outputMatchers {
				// a pack can match several Output
				if matcher != nil && matcher.Match(pack) {
					foundMatch = true

					pack.incRef()
					pack.diagnostics.AddStamp(matcher.runner)
					matcher.InChan() <- pack
				}
			}

			// got pack from Input, now dispatch
			// for each target, pack will inc ref count
			// and the router will dec ref count only once
			for _, matcher = range this.filterMatchers {
				// a pack can match several Filter
				if matcher != nil && matcher.Match(pack) {
					foundMatch = true

					pack.incRef()
					pack.diagnostics.AddStamp(matcher.runner)
					matcher.InChan() <- pack
				}
			}

			if !foundMatch {
				// Maybe we closed all filter/output inChan, but there
				// still exits some remnant packs in router.hub
				log.Warn("no match: %+v", pack)
			}

			// never forget this!
			pack.Recycle()
		}
	}

	log.Trace("Router shutdown.")
}

func (this *messageRouter) removeMatcher(matcher *matcher, matchers []*matcher) {
	for idx, m := range matchers {
		if m == matcher {
			log.Trace("closing matcher for %s", m.runner.Name())

			// in golang, close means we can no longer send to that chan
			// but consumers can still recv from the chan
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

func (this *routerStats) update(pack *PipelinePack) {
	atomic.AddInt64(&this.TotalProcessedMsgN, 1)
	atomic.AddInt32(&this.PeriodProcessedMsgN, 1)

	if pack.input {
		atomic.AddInt64(&this.TotalInputMsgN, 1)
		atomic.AddInt32(&this.PeriodInputMsgN, 1)
		atomic.AddInt64(&this.TotalInputBytes, int64(pack.Payload.Length()))
		atomic.AddInt64(&this.PeriodInputBytes, int64(pack.Payload.Length()))
	}
}

func (this *routerStats) resetPeriodCounters() {
	this.PeriodProcessedBytes = int64(0)
	this.PeriodInputBytes = int64(0)
	this.PeriodInputMsgN = int32(0)
	this.PeriodProcessedMsgN = int32(0)
	this.PeriodMaxMsgBytes = int64(0)
}

func (this *routerStats) render(elapsed int) {
	log.Trace("Total:%10s %10s speed:%6s/s %10s/s max: %s/%s",
		gofmt.Comma(this.TotalProcessedMsgN),
		gofmt.ByteSize(this.TotalProcessedBytes),
		gofmt.Comma(int64(this.PeriodProcessedMsgN/int32(elapsed))),
		gofmt.ByteSize(this.PeriodProcessedBytes/int64(elapsed)),
		gofmt.ByteSize(this.PeriodMaxMsgBytes),
		gofmt.ByteSize(this.TotalMaxMsgBytes))
	log.Trace("Input:%10s %10s speed:%6s/s %10s/s",
		gofmt.Comma(int64(this.PeriodInputMsgN)),
		gofmt.ByteSize(this.PeriodInputBytes),
		gofmt.Comma(int64(this.PeriodInputMsgN/int32(elapsed))),
		gofmt.ByteSize(this.PeriodInputBytes/int64(elapsed)))
}
