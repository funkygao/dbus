package engine

import (
	"sync"

	"github.com/funkygao/go-metrics"
)

type routerMetrics struct {
	l sync.RWMutex

	m map[string]metrics.Meter   // key is Packet.Ident
	c map[string]metrics.Counter // key is Packet.Ident
}

func newMetrics() *routerMetrics {
	return &routerMetrics{
		m: make(map[string]metrics.Meter, 10),
		c: make(map[string]metrics.Counter, 10),
	}
}

func (m *routerMetrics) Update(pack *Packet) {
	m.l.RLock()
	_, present := m.m[pack.Ident]
	m.l.RUnlock()
	if !present {
		m.l.Lock()
		m.m[pack.Ident] = metrics.NewRegisteredMeter(pack.Ident, metrics.DefaultRegistry)
		m.c[pack.Ident] = metrics.NewRegisteredCounter(pack.Ident, metrics.DefaultRegistry)
		m.l.Unlock()
	}

	m.m[pack.Ident].Mark(1)
	m.c[pack.Ident].Inc(1)
}
