package engine

import (
	"github.com/funkygao/go-metrics"
)

// routerMetrics is not thread safe because it has no concurrency.
type routerMetrics struct {
	m map[string]metrics.Meter // key is Packet.Ident
}

func newMetrics() *routerMetrics {
	return &routerMetrics{
		m: make(map[string]metrics.Meter, 10),
	}
}

func (m *routerMetrics) Update(pack *Packet) {
	if _, present := m.m[pack.Ident]; !present {
		m.m[pack.Ident] = metrics.NewRegisteredMeter(pack.Ident, metrics.DefaultRegistry)
	}

	m.m[pack.Ident].Mark(1)
}
