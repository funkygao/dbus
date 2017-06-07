// dbusd is the dbus daemon.
package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/funkygao/dbus"
	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gafka/diagnostics/agent"
	"github.com/funkygao/log4go"

	// bootstrap plugins
	_ "github.com/funkygao/dbus/plugins"
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

	fmt.Fprint(os.Stderr, logo[1:])

	globals := engine.DefaultGlobals()
	globals.Debug = options.debug
	globals.RPCPort = options.rpcPort
	globals.APIPort = options.apiPort
	globals.RouterTrack = options.routerTrack
	globals.InputRecyclePoolSize = options.inputPoolSize
	globals.FilterRecyclePoolSize = options.filterPoolSize
	globals.HubChanSize = options.hubPoolSize
	globals.PluginChanSize = options.pluginPoolSize
	globals.ClusterEnabled = options.clusterEnable
	globals.Zone = options.zone
	globals.Cluster = options.cluster

	if !options.validateConf && len(options.visualizeFile) == 0 {
		// daemon mode
		log4go.Info("dbus[%s@%s] starting for {zone:%s cluster:%s}",
			dbus.Revision, dbus.Version, options.zone, options.cluster)

		agent.HttpAddr = options.pprofAddr
		log4go.Info("pprof agent ready on %s", agent.Start())
		go func() {
			log4go.Error("%s", <-agent.Errors)
		}()
	}

	t0 := time.Now()
	var err error
	for {
		e := engine.New(globals).LoadFrom(options.configPath)

		if options.visualizeFile != "" {
			e.ExportDiagram(options.visualizeFile)
			return
		}

		if options.validateConf {
			fmt.Println("ok")
			return
		}

		if err = e.ServeForever(); err != nil {
			// e,g. SIGTERM received
			break
		}
	}

	log4go.Info("dbus[%s@%s] %s, bye!", dbus.Revision, dbus.Version, time.Since(t0))
	log4go.Close()
}
