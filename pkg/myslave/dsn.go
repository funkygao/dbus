package myslave

import (
	"net"
	"net/url"
	"strconv"
	"strings"

	parser "github.com/funkygao/dbus/pkg/dsn"
)

// ParseDSN parse mysql DSN(data source name).
// The DSN is in the form of mysql:zone://user:pass@host:port/db1,db2,...,dbn
// The zone is used for zk checkpoint.
func ParseDSN(dsn string) (zone, host string, port uint16, username, passwd string, dbs []string, err error) {
	var scheme string
	if scheme, dsn, err = parser.Parse(dsn); err != nil {
		return
	} else if scheme != "mysql" {
		err = parser.ErrIllegalDSN
		return
	}

	var u *url.URL
	if u, err = url.Parse(dsn); err != nil {
		return
	}

	var p string
	if host, p, err = net.SplitHostPort(u.Host); err != nil {
		return
	}

	username = u.User.Username()
	passwd, _ = u.User.Password()
	var portInt int
	if portInt, err = strconv.Atoi(p); err != nil {
		return
	}
	port = uint16(portInt)

	databases := strings.Split(strings.TrimPrefix(u.Path, "/"), ",")
	dbs = make([]string, 0, len(databases))
	for _, db := range databases {
		if db = strings.TrimSpace(db); db != "" {
			dbs = append(dbs, db)
		}
	}

	zone = u.Scheme
	return
}
