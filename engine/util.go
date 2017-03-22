package engine

import (
	"strings"
)

func parseConfigPath(path string) (zkSvr string, realPath string) {
	if !strings.Contains(path, ":") {
		realPath = path
		return
	}

	tuples := strings.SplitN(path, "/", 2)
	zkSvr = tuples[0]
	realPath = "/" + tuples[1]
	return
}
