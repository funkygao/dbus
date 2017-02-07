package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/funkygao/log4go"
)

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

	prefix := fmt.Sprintf("[%d] ", os.Getpid())
	log.SetOutput(logWriter)
	log.SetFlags(logOptions)
	log.SetPrefix(prefix)

	return log.New(logWriter, prefix, logOptions)
}

func setupLogging() {
	log4go.SetLevel(log4go.ToLogLevel(options.loglevel, log4go.TRACE))
}
