package myslave

import (
	"fmt"

	"github.com/funkygao/go-metrics"
)

type slaveMetrics struct {
	Lag metrics.Gauge

	TPS    metrics.Meter
	Events metrics.Meter
}

func newMetrics(prefix string) *slaveMetrics {
	m := &slaveMetrics{}

	m.Lag = metrics.NewRegisteredGauge(fmt.Sprintf("%s.lag", prefix), metrics.DefaultRegistry)
	m.TPS = metrics.NewRegisteredMeter(fmt.Sprintf("%s.tps", prefix), metrics.DefaultRegistry)
	m.Events = metrics.NewRegisteredMeter(fmt.Sprintf("%s.events", prefix), metrics.DefaultRegistry)
	return m
}
