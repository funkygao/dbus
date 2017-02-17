package myslave

import (
	"net"
	"strconv"
	"strings"
)

func configMasterAddr(addr string) (masterAddr string, host string, port uint16, schema string) {
	slashIdx := strings.LastIndex(addr, "/")
	if slashIdx != -1 {
		if slashIdx < len(addr)-1 {
			schema = addr[slashIdx+1:]
		}
		addr = addr[:slashIdx]
	}

	masterAddr = addr
	h, p, err := net.SplitHostPort(masterAddr)
	if err != nil {
		panic(err)
	}

	host = h
	pt, err := strconv.Atoi(p)
	if err != nil {
		panic(err)
	}
	port = uint16(pt)

	return
}
