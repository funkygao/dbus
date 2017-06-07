package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/funkygao/dbus"
	"github.com/funkygao/gafka/ctx"
)

var (
	options struct {
		debug   bool
		zone    string
		cluster string

		configPath    string
		validateConf  bool
		showversion   bool
		visualizeFile string
		pprofAddr     string
		lockfile      string
		routerTrack   bool
		clusterEnable bool

		logfile  string
		loglevel string

		inputPoolSize  int
		filterPoolSize int
		hubPoolSize    int
		pluginPoolSize int

		zrootCheckpoint string
		zrootCluster    string
		zrootConfig     string

		apiPort int
		rpcPort int
	}
)

const (
	usage = `dbusd - Distributed DataBus

Flags:
`
)

func parseFlags() {
	iPool := 3000
	fPool := iPool * 15 / 10
	hPool := 3 * iPool
	pPool := iPool
	flag.StringVar(&options.configPath, "conf", "", "main config path: file or zk accepted")
	flag.BoolVar(&options.validateConf, "validate", false, "validate config file and exit")
	flag.StringVar(&options.logfile, "logfile", "", "master log file path, default stdout")
	flag.StringVar(&options.loglevel, "loglevel", "trace", "log level")
	flag.BoolVar(&options.showversion, "version", false, "show version and exit")
	flag.BoolVar(&options.debug, "debug", false, "debug mode")
	flag.StringVar(&options.pprofAddr, "pprof", ":10120", "pprof agent listen address")
	flag.BoolVar(&options.routerTrack, "routerstat", true, "track router metrics")
	flag.IntVar(&options.inputPoolSize, "ipool", iPool, "input recycle pool size")
	flag.IntVar(&options.filterPoolSize, "fpool", fPool, "filter recycle pool size")
	flag.BoolVar(&options.clusterEnable, "cluster", false, "enable cluster feature")
	flag.IntVar(&options.hubPoolSize, "hpool", hPool, "hub pool size")
	flag.IntVar(&options.pluginPoolSize, "ppool", pPool, "plugin pool size")
	flag.IntVar(&options.rpcPort, "rpc", 9877, "rpc server port")
	flag.IntVar(&options.apiPort, "api", 9876, "api server port")
	flag.StringVar(&options.zone, "z", ctx.DefaultZone(), "zone")
	flag.StringVar(&options.cluster, "c", "", "")
	flag.StringVar(&options.zrootCheckpoint, "rootcheckpoint", "", "checkpoint znode root")
	flag.StringVar(&options.zrootCluster, "rootcluster", "", "cluster znode root")
	flag.StringVar(&options.zrootConfig, "rootconfig", "", "config znode root")
	flag.StringVar(&options.visualizeFile, "dump", "", "visualize the pipleline to a png file. graphviz must be installed")
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(options.cluster) == 0 {
		panic("-c required")
	}
}

func showVersionAndExit() {
	fmt.Fprintf(os.Stderr, "%s %s (%s)\n", os.Args[0], dbus.Version, dbus.Revision)
	fmt.Fprintf(os.Stderr, "Built with %s %s for %s/%s at %s by %s\n",
		runtime.Compiler, runtime.Version(), runtime.GOOS, runtime.GOARCH, dbus.BuildDate, dbus.BuildUser)
	os.Exit(0)
}
