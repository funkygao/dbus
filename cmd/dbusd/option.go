package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

func parseFlags() {
	flag.BoolVar(&options.verbose, "v", false, "verbose")
	flag.BoolVar(&options.veryVerbose, "vv", false, "very verbose")
	flag.StringVar(&options.configfile, "conf", "etc/engine.als.cf", "main config file")
	flag.StringVar(&options.logfile, "log", "", "master log file path, default stdout")
	flag.StringVar(&options.lockfile, "lockfile", "var/dpiped.lock", "lockfile path")
	flag.BoolVar(&options.showversion, "version", false, "show version and exit")
	flag.BoolVar(&options.debug, "debug", false, "debug mode")
	flag.IntVar(&options.tick, "tick", 60*10, "tick interval in seconds to report sys stat")
	flag.BoolVar(&options.dryrun, "dryrun", false, "dry run")
	flag.StringVar(&options.cpuprof, "cpuprof", "", "cpu profiling file")
	flag.StringVar(&options.memprof, "memprof", "", "memory profiling file")
	flag.Usage = showUsage

	flag.Parse()

	if options.veryVerbose {
		options.debug = true
	}
	if options.debug {
		options.verbose = true
	}

	if options.tick <= 0 {
		panic("tick must be possitive")
	}
}

func showUsage() {
	fmt.Fprint(os.Stderr, USAGE)
	flag.PrintDefaults()
}

func newLogger() *log.Logger {
	var logWriter io.Writer = os.Stdout // default log writer
	var err error
	if options.logfile != "" {
		logWriter, err = os.OpenFile(options.logfile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			panic(err)
		}
	}

	logOptions := log.Ldate | log.Ltime | log.Lshortfile
	if options.debug {
		logOptions |= log.Lmicroseconds
	}

	prefix := fmt.Sprintf("[%d]", os.Getpid())
	log.SetOutput(logWriter)
	log.SetFlags(logOptions)
	log.SetPrefix(prefix)

	return log.New(logWriter, prefix, logOptions)
}
