package myslave

import (
	"strings"

	"github.com/funkygao/gafka/telemetry"
	"github.com/funkygao/go-metrics"
)

type slaveMetrics struct {
	Lag metrics.Gauge

	TPS    metrics.Meter
	Events metrics.Meter
}

func newMetrics(name string) *slaveMetrics {
	tag := telemetry.Tag(strings.Replace(name, ".", "_", -1), "", "v1")
	return &slaveMetrics{
		Lag:    metrics.NewRegisteredGauge(tag+"mysql.binlog.lag", metrics.DefaultRegistry),
		TPS:    metrics.NewRegisteredMeter(tag+"mysql.binlog.tps", metrics.DefaultRegistry),
		Events: metrics.NewRegisteredMeter(tag+"mysql.binlog.evt", metrics.DefaultRegistry),
	}
}
