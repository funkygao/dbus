package main

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/funkygao/dbus"
	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gafka/diagnostics/agent"
	"github.com/funkygao/log4go"

	// bootstrap plugins
	_ "github.com/funkygao/dbus/plugins/filter"
	_ "github.com/funkygao/dbus/plugins/input"
	_ "github.com/funkygao/dbus/plugins/output"
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
	globals.RecyclePoolSize = options.mpoolSize
	globals.PluginChanSize = options.ppoolSize

	e := engine.New(globals).
		LoadConfigFile(options.configfile)

	if options.visualizeFile != "" {
		e.ExportDiagram(options.visualizeFile)
		return
	}

	log4go.Info("dbus[%s@%s] starting", dbus.BuildID, dbus.Version)

	agent.HttpAddr = ":10120" // FIXME security issue
	log4go.Info("pprof agent ready on %s", agent.Start())

	t0 := time.Now()
	e.ServeForever()

	log4go.Info("dbus[%s@%s] %s, bye!", dbus.BuildID, dbus.Version, time.Since(t0))
	log4go.Close()
}
