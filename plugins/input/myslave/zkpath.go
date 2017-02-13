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
	return fmt.Sprintf("%s/ids/%s", posPath(masterAddr), myNode())
}

func myNode() string {
	host, _ := os.Hostname()
	return fmt.Sprintf("%s-%d", host, os.Getpid())
}
