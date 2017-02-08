package main

import (
	"fmt"
	"runtime/debug"

	"github.com/funkygao/dbus/engine"
	_ "github.com/funkygao/dbus/plugins/filter"
	_ "github.com/funkygao/dbus/plugins/input"
	_ "github.com/funkygao/dbus/plugins/output"
	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gafka/diagnostics/agent"
	"github.com/funkygao/log4go"
)

func init() {
	parseFlags()

	if options.showversion {
		showVersionAndExit()
	}

	setupLogging()

	ctx.LoadFromHome()
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			debug.PrintStack()
		}
	}()

	globals := engine.DefaultGlobals()
	globals.Debug = options.debug
	globals.Verbose = options.verbose
	globals.VeryVerbose = options.veryVerbose
	globals.DryRun = options.dryrun
	globals.Logger = newLogger()

	e := engine.New(globals).
		LoadConfigFile(options.configfile)

	if options.visualizeFile != "" {
		e.ExportDiagram(options.visualizeFile)
		return
	}

	agent.HttpAddr = ":10120" // FIXME security issue
	log4go.Info("pprof agent ready on %s", agent.Start())

	e.ServeForever()
}
