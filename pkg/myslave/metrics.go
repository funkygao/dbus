package myslave

import (
	"fmt"
	"strings"

	"github.com/funkygao/gafka/telemetry"
	"github.com/funkygao/go-metrics"
)

type slaveMetrics struct {
	Lag metrics.Gauge

	TPS    metrics.Meter
	Events metrics.Meter
}

func newMetrics(host string, port uint16) *slaveMetrics {
	tag := telemetry.Tag(strings.Replace(host, ".", "_", -1), fmt.Sprintf("%d", port), "v1")
	return &slaveMetrics{
		Lag:    metrics.NewRegisteredGauge(tag+"mysql.binlog.lag", metrics.DefaultRegistry),
		TPS:    metrics.NewRegisteredMeter(tag+"mysql.binlog.tps", metrics.DefaultRegistry),
		Events: metrics.NewRegisteredMeter(tag+"mysql.binlog.evt", metrics.DefaultRegistry),
	}
}
