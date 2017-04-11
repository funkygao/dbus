package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/funkygao/dbus"
	"github.com/funkygao/log4go"
)

func setupLogging() {
	// keep the world silent
	log.SetOutput(ioutil.Discard)

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

	f, err := os.OpenFile("panic", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		panic(err)
	}
	syscall.Dup2(int(f.Fd()), int(os.Stdout.Fd()))
	syscall.Dup2(int(f.Fd()), int(os.Stderr.Fd()))
	fmt.Fprintf(os.Stderr, "\n%s %s (build: %s)\n===================\n",
		time.Now(), dbus.Version, dbus.Revision)
}
