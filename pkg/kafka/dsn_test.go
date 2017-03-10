package kafka

import (
	"testing"

	"github.com/funkygao/assert"
)

func TestParseDSN(t *testing.T) {
	dsn := "kafka://prod:trade/orders"
	zone, cluster, topic, err := ParseDSN(dsn)
	assert.Equal(t, nil, err)
	assert.Equal(t, "prod", zone)
	assert.Equal(t, "trade", cluster)
	assert.Equal(t, "orders", topic)
}
