package myslave

import (
	"net"
	"net/url"
	"strconv"
	"strings"
)

const mysqlSchemePrefix = "mysql://"

// parseDSN parse mysql DSN(data source name).
func parseDSN(dsn string) (host string, port uint16, username, passwd string, dbs []string, err error) {
	if !strings.HasPrefix(dsn, mysqlSchemePrefix) {
		dsn = mysqlSchemePrefix + dsn
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

	return
}
