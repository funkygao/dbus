package myslave

import (
	"strings"

	"github.com/funkygao/gafka/telemetry"
	"github.com/funkygao/go-metrics"
)

type slaveMetrics struct {
	tag string

	Lag    metrics.Gauge
	TPS    metrics.Meter
	Events metrics.Meter
}

func newMetrics(name string) *slaveMetrics {
	tag := telemetry.Tag(strings.Replace(name, ".", "_", -1), "", "v1")
	return &slaveMetrics{
		tag:    tag,
		Lag:    metrics.NewRegisteredGauge(tag+"mysql.binlog.lag", metrics.DefaultRegistry),
		TPS:    metrics.NewRegisteredMeter(tag+"mysql.binlog.tps", metrics.DefaultRegistry),
		Events: metrics.NewRegisteredMeter(tag+"mysql.binlog.evt", metrics.DefaultRegistry),
	}
}

func (m *slaveMetrics) Close() {
	// TODO flush metrics
	metrics.Unregister(m.tag + "mysql.binlog.lag")
	metrics.Unregister(m.tag + "mysql.binlog.tps")
	metrics.Unregister(m.tag + "mysql.binlog.evt")
}
