package kafka

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

// ParseDSN parse the kafka DSN which is in the form of: zone://cluster/topic#partition.
func ParseDSN(dsn string) (zone, cluster, topic string, partitionID int32, err error) {
	var u *url.URL
	if u, err = url.Parse(dsn); err != nil {
		return
	}

	partitionID = InvalidPartitionID
	cluster = u.Host
	zone = u.Scheme
	topic = strings.TrimPrefix(u.Path, "/")
	if zone == "" || cluster == "" {
		err = errors.New("empty zone or cluster")
		return
	}

	if len(u.Fragment) > 0 {
		var p int
		if p, err = strconv.Atoi(u.Fragment); err != nil {
			return
		}
		partitionID = int32(p)
	}

	return
}
