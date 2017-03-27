package myslave

import (
	"fmt"
	"os"
	"path"
)

var (
	rootPath  = "/dbus"
	slaveRoot = path.Join(rootPath, "myslave")
)

func posPath(masterAddr string) string {
	return fmt.Sprintf("%s/%s", slaveRoot, masterAddr)
}

func masterPath(masterAddr string) string {
	return fmt.Sprintf("%s/owner", posPath(masterAddr))
}

func myNodePath(masterAddr string) string {
	return fmt.Sprintf("%s/ids/%s", posPath(masterAddr), myNode())
}

func myNode() string {
	host, _ := os.Hostname()
	return fmt.Sprintf("%s-%d", host, os.Getpid())
}
