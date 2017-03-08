package engine

import (
	//"github.com/funkygao/gafka/telemetry"
	"github.com/funkygao/go-metrics"
)

type routerMetrics struct {
	in  map[string]metrics.Meter
	out map[string]metrics.Meter
}

func newMetrics() *routerMetrics {
	return &routerMetrics{}
}

func (m *routerMetrics) update(pack *Packet) {

}
