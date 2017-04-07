package kafka

type zoneCluster struct {
	zone    string
	cluster string
}

// TODO rename
type topicPartitions struct {
	zc  zoneCluster
	tps map[string][]int32 // topic:partitions
}

func parseDSNs(dsns []string) ([]topicPartitions, error) {
	var m = make(map[zoneCluster]map[string][]int32)
	for _, dsn := range dsns {
		zone, cluster, topic, partitionID, err := ParseDSN(dsn)
		if err != nil {
			return nil, err
		}

		zc := zoneCluster{zone: zone, cluster: cluster}
		if _, present := m[zc]; !present {
			m[zc] = make(map[string][]int32)
		}
		if _, present := m[zc][topic]; !present {
			m[zc][topic] = make([]int32, 0)
		}

		// TODO check dup partitionID
		m[zc][topic] = append(m[zc][topic], partitionID)
	}

	var r []topicPartitions
	for zc, tps := range m {
		r = append(r, topicPartitions{zc: zc, tps: tps})
	}

	return r, nil
}
