package myslave

import (
	"fmt"
	"os"
)

func posPath(masterAddr string) string {
	return fmt.Sprintf("/dbus/myslave/%s", masterAddr)
}

func masterPath(masterAddr string) string {
	return fmt.Sprintf("%s/master", posPath(masterAddr))
}

func myNodePath(masterAddr string) string {
	host, _ := os.Hostname()
	return fmt.Sprintf("%s/nodes/%s-%d", posPath(masterAddr), host, os.Getpid())
}
