package myslave

import (
	"net"
	"net/url"
	"strconv"
	"strings"
)

func parseDSN(dsn string) (host string, port uint16, username, passwd string, err error) {
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

	return
}
