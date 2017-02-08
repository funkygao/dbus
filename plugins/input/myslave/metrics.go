package myslave

import (
	"fmt"

	"github.com/funkygao/go-metrics"
)

type metrics1 struct {
	prefix string

	Lag metrics.Gauge
	TPS metrics.Meter
}

func newMetrics(prefix string) *metrics1 {
	m := &metrics1{
		prefix: prefix,
	}
	m.Lag = metrics.NewRegisteredGauge(fmt.Sprintf("%s.lag", prefix), metrics.DefaultRegistry)
	m.TPS = metrics.NewRegisteredMeter(fmt.Sprintf("%s.tps", prefix), metrics.DefaultRegistry)
	return m
}
