package kafka

import (
	"net"
	"net/url"
	"strings"
)

const kafkaSchemePrefix = "kafka://"

// ParseDSN parse the kafka DSN which is in the form of: kafka://zone:cluster/topic.
func ParseDSN(dsn string) (zone, cluster, topic string, err error) {
	if !strings.HasPrefix(dsn, kafkaSchemePrefix) {
		dsn = kafkaSchemePrefix + dsn
	}

	var u *url.URL
	if u, err = url.Parse(dsn); err != nil {
		return
	}

	if zone, cluster, err = net.SplitHostPort(u.Host); err != nil {
		return
	}

	topic = strings.TrimPrefix(u.Path, "/")
	return
}
