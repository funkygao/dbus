package main

import (
	"fmt"
	"log"
	"runtime/debug"

	"github.com/funkygao/dbus/engine"
	_ "github.com/funkygao/dbus/plugins/filter"
	_ "github.com/funkygao/dbus/plugins/input"
	_ "github.com/funkygao/dbus/plugins/output"
	"github.com/funkygao/gafka/diagnostics/agent"
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
	globals.Logger = newLogger()
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			debug.PrintStack()
		}
	}()

	log.Printf("pprof ready on %s", agent.Start())

	e := engine.New(globals).
		LoadConfigFile(options.configfile)
	if options.visualizeFile != "" {
		e.ExportDiagram(options.visualizeFile)
		return
	}

	e.ServeForever()
}
