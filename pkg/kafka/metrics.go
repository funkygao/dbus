package kafka

import (
	"strings"

	"github.com/funkygao/gafka/telemetry"
	"github.com/funkygao/go-metrics"
)

type producerMetrics struct {
	name string

	asyncSend metrics.Meter
	asyncOk   metrics.Meter
	asyncFail metrics.Meter

	syncOk   metrics.Meter
	syncFail metrics.Meter
}

func newMetrics(name string) *producerMetrics {
	tag := telemetry.Tag(strings.Replace(name, ".", "_", -1), "", "")
	return &producerMetrics{
		name:      name,
		asyncSend: metrics.NewRegisteredMeter(tag+"dbus.kafka.async.send", metrics.DefaultRegistry),
		asyncOk:   metrics.NewRegisteredMeter(tag+"dbus.kafka.async.ok", metrics.DefaultRegistry),
		asyncFail: metrics.NewRegisteredMeter(tag+"dbus.kafka.async.fail", metrics.DefaultRegistry),
		syncOk:    metrics.NewRegisteredMeter(tag+"dbus.kafka.sync.ok", metrics.DefaultRegistry),
		syncFail:  metrics.NewRegisteredMeter(tag+"dbus.kafka.sync.fail", metrics.DefaultRegistry),
	}
}
