package kafka

import (
	"errors"
	"net/url"
	"strings"
)

// ParseDSN parse the kafka DSN which is in the form of: zone://cluster/topic.
func ParseDSN(dsn string) (zone, cluster, topic string, err error) {
	var u *url.URL
	if u, err = url.Parse(dsn); err != nil {
		return
	}

	cluster = u.Host
	zone = u.Scheme
	topic = strings.TrimPrefix(u.Path, "/")
	if zone == "" || cluster == "" {
		err = errors.New("empty zone or cluster")
		return
	}
	return
}
