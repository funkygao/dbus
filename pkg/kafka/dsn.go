package kafka

import (
	"errors"
	"net/url"
	"strconv"
	"strings"

	parser "github.com/funkygao/dbus/pkg/dsn"
)

// ParseDSN parse the kafka DSN which is in the form of: kafka:zone://cluster/topic#partition.
func ParseDSN(dsn string) (zone, cluster, topic string, partitionID int32, err error) {
	var scheme string
	if scheme, dsn, err = parser.Parse(dsn); err != nil {
		return
	} else if scheme != "kafka" {
		err = parser.ErrIllegalDSN
		return
	}

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
		if p < 0 {
			err = errors.New("partition id < 0")
			return
		}
		partitionID = int32(p)
	}

	return
}
