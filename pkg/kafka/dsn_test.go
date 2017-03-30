package kafka

import (
	"testing"

	"github.com/funkygao/assert"
)

func TestParseDSN(t *testing.T) {
	dsn := "prod://trade/orders"
	zone, cluster, topic, partitionID, err := ParseDSN(dsn)
	assert.Equal(t, nil, err)
	assert.Equal(t, "prod", zone)
	assert.Equal(t, "trade", cluster)
	assert.Equal(t, "orders", topic)
	assert.Equal(t, InvalidPartitionID, partitionID)

	dsn = "prod://trade/orders#"
	zone, cluster, topic, partitionID, err = ParseDSN(dsn)
	assert.Equal(t, nil, err)
	assert.Equal(t, "prod", zone)
	assert.Equal(t, "trade", cluster)
	assert.Equal(t, "orders", topic)
	assert.Equal(t, int32(-1), partitionID)

	dsn = "prod://trade/orders#1"
	zone, cluster, topic, partitionID, err = ParseDSN(dsn)
	assert.Equal(t, nil, err)
	assert.Equal(t, "prod", zone)
	assert.Equal(t, "trade", cluster)
	assert.Equal(t, "orders", topic)
	assert.Equal(t, int32(1), partitionID)

	dsn = "prod://trade/orders#0"
	zone, cluster, topic, partitionID, err = ParseDSN(dsn)
	assert.Equal(t, nil, err)
	assert.Equal(t, "prod", zone)
	assert.Equal(t, "trade", cluster)
	assert.Equal(t, "orders", topic)
	assert.Equal(t, int32(0), partitionID)

	dsn = "prod://trade/orders#invalid"
	zone, cluster, topic, partitionID, err = ParseDSN(dsn)
	assert.Equal(t, InvalidPartitionID, partitionID)
	assert.Equal(t, true, err != nil)

	dsn = "prod://trade/orders#-9"
	zone, cluster, topic, partitionID, err = ParseDSN(dsn)
	assert.Equal(t, InvalidPartitionID, partitionID)
	assert.Equal(t, true, err != nil)

	// empty zone raises err
	dsn = "trade/orders"
	zone, cluster, topic, partitionID, err = ParseDSN(dsn)
	assert.Equal(t, false, err == nil)

	// empty cluster raises err
	dsn = "uat:///orders"
	zone, cluster, topic, partitionID, err = ParseDSN(dsn)
	assert.Equal(t, "uat", zone)
	assert.Equal(t, "", cluster)
	assert.Equal(t, "orders", topic)
	assert.Equal(t, false, err == nil)
	assert.Equal(t, int32(-1), partitionID)
}
