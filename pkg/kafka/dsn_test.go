package kafka

import (
	"testing"

	"github.com/funkygao/assert"
)

func TestParseDSN(t *testing.T) {
	dsn := "prod://trade/orders"
	zone, cluster, topic, err := ParseDSN(dsn)
	assert.Equal(t, nil, err)
	assert.Equal(t, "prod", zone)
	assert.Equal(t, "trade", cluster)
	assert.Equal(t, "orders", topic)

	// empty zone raises err
	dsn = "trade/orders"
	zone, cluster, topic, err = ParseDSN(dsn)
	assert.Equal(t, false, err == nil)

	// empty cluster raises err
	dsn = "uat:///orders"
	zone, cluster, topic, err = ParseDSN(dsn)
	assert.Equal(t, "uat", zone)
	assert.Equal(t, "", cluster)
	assert.Equal(t, "orders", topic)
	assert.Equal(t, false, err == nil)
}
