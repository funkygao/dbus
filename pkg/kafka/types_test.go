package kafka

import (
	"testing"

	"github.com/funkygao/assert"
)

func TestParseDSNs(t *testing.T) {
	dsns := []string{
		"test://trade/orders#0",
		"prod://user/teacher#2",
		"prod://user/teacher#3",
		"prod://user/foobar#1",
	}

	tps, err := parseDSNs(dsns)
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(tps))                                                  // 2 zone.cluster
	assert.Equal(t, 1, len(getByZoneCluster(tps, "test", "trade").tps))           // prod.trade has 1 topic
	assert.Equal(t, 2, len(getByZoneCluster(tps, "prod", "user").tps))            // prod.user has 2 topics
	assert.Equal(t, 1, len(getByZoneCluster(tps, "prod", "user").tps["foobar"]))  // foobar has 1 partition
	assert.Equal(t, 2, len(getByZoneCluster(tps, "prod", "user").tps["teacher"])) // teacher has 2 partitions
}

func getByZoneCluster(tps []topicPartitions, zone, cluster string) *topicPartitions {
	for _, tp := range tps {
		if tp.zc.zone == zone && tp.zc.cluster == cluster {
			return &tp
		}
	}

	return nil
}
