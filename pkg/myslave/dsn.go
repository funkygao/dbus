package myslave

import (
	"net"
	"net/url"
	"strconv"
	"strings"
)

// parseDSN parse mysql DSN(data source name).
func parseDSN(dsn string) (host string, port uint16, username, passwd string, dbs []string, err error) {
	if !strings.HasPrefix(dsn, "mysql://") {
		dsn = "mysql://" + dsn
	}

	var u *url.URL
	u, err = url.Parse(dsn)
	if err != nil {
		return
	}

	var p string
	host, p, err = net.SplitHostPort(u.Host)
	if err != nil {
		return
	}

	username = u.User.Username()
	passwd, _ = u.User.Password()
	var portInt int
	portInt, err = strconv.Atoi(p)
	if err != nil {
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

	return
}
