package main

import (
	"fmt"
	"os"
	"runtime"
)

const (
	VERSION = "v0.5.2.stable"
	AUTHOR  = "funky.gao@gmail.com"
)

func showVersionAndExit() {
	fmt.Fprintf(os.Stderr, "%s %s (build: %s)\n", os.Args[0], VERSION, BuildID)
	fmt.Fprintf(os.Stderr, "Built with %s %s for %s/%s\n",
		runtime.Compiler, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	os.Exit(0)
}
