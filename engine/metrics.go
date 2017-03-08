package engine

import (
	"strings"

	"github.com/funkygao/gafka/telemetry"
	"github.com/funkygao/go-metrics"
)

type routerMetrics struct {
	asyncSend metrics.Meter
	asyncOk   metrics.Meter
	asyncFail metrics.Meter

	syncOk   metrics.Meter
	syncFail metrics.Meter
}

func newMetrics() *routerMetrics {
	name := ""
	tag := telemetry.Tag(strings.Replace(name, ".", "_", -1), "", "")
	return &routerMetrics{
		asyncSend: metrics.NewRegisteredMeter(tag+"dbus.kafka.async.send", metrics.DefaultRegistry),
		asyncOk:   metrics.NewRegisteredMeter(tag+"dbus.kafka.async.ok", metrics.DefaultRegistry),
		asyncFail: metrics.NewRegisteredMeter(tag+"dbus.kafka.async.fail", metrics.DefaultRegistry),
		syncOk:    metrics.NewRegisteredMeter(tag+"dbus.kafka.sync.ok", metrics.DefaultRegistry),
		syncFail:  metrics.NewRegisteredMeter(tag+"dbus.kafka.sync.fail", metrics.DefaultRegistry),
	}
}

func (m *routerMetrics) update(pack *Packet) {

}
