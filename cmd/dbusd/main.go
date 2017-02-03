package main

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/funkygao/dbus/engine"
	_ "github.com/funkygao/dbus/plugins" // trigger RegisterPlugin(s)
	"github.com/funkygao/gafka/diagnostics"
)

func init() {
	parseFlags()

	if options.showversion {
		showVersionAndExit()
	}

	globals = engine.DefaultGlobals()
	globals.Debug = options.debug
	globals.Verbose = options.verbose
	globals.VeryVerbose = options.veryVerbose
	globals.DryRun = options.dryrun
	globals.TickerLength = options.tick
	globals.Logger = newLogger()
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			debug.PrintStack()
		}
	}()

	ticker := time.NewTicker(time.Second * time.Duration(options.tick))
	defer ticker.Stop()

	go diagnostics.Start()

	engine.New(globals).
		LoadConfigFile(options.configfile).
		ServeForever()
}
