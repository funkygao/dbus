package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/funkygao/dbus"
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
	level := log4go.ToLogLevel(options.loglevel, log4go.TRACE)
	log4go.SetLevel(level)
	if options.logfile == "" {
		// stdout
		return
	}

	log4go.DeleteFilter("stdout")
	rotateEnabled, discardWhenDiskFull := true, false
	filer := log4go.NewFileLogWriter(options.logfile, rotateEnabled, discardWhenDiskFull, 0644)
	filer.SetFormat("[%d %T] [%L] (%S) %M")
	filer.SetRotateKeepDuration(time.Hour * 24 * 30)
	filer.SetRotateLines(0)
	filer.SetRotateDaily(true)
	log4go.AddFilter("file", level, filer)

	f, err := os.OpenFile("panic", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	syscall.Dup2(int(f.Fd()), int(os.Stdout.Fd()))
	syscall.Dup2(int(f.Fd()), int(os.Stderr.Fd()))
	fmt.Fprintf(os.Stderr, "\n%s %s (build: %s)\n===================\n",
		time.Now(),
		dbus.Version, dbus.BuildID)
}
