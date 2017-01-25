package engine

import (
	"log"
	"os"
	"regexp"
	"syscall"
	"time"
)

const (
	RELOAD  = "reload"
	STOP    = "stop"
	SIGUSR1 = "user1"
	SIGUSR2 = "user2"
)

var (
	availablePlugins = make(map[string]func() Plugin)
	pluginTypeRegex  = regexp.MustCompile("^.*(Filter|Input|Output)$")

	Globals func() *GlobalConfigStruct
)

// Struct for holding global pipeline config values.
type GlobalConfigStruct struct {
	*log.Logger

	StartedAt       time.Time
	Stopping        bool
	Debug           bool
	Verbose         bool
	VeryVerbose     bool
	DryRun          bool
	RecyclePoolSize int
	PluginChanSize  int
	TickerLength    int

	MaxMsgLoops int
	MaxPackIdle time.Duration

	sigChan chan os.Signal
}

func (this *GlobalConfigStruct) Shutdown() {
	this.Kill(syscall.SIGINT)
}

func (this *GlobalConfigStruct) Kill(sig os.Signal) {
	go func(s os.Signal) {
		this.sigChan <- s
	}(sig)
}

func DefaultGlobals() *GlobalConfigStruct {
	idle, _ := time.ParseDuration("2m")
	return &GlobalConfigStruct{
		Debug:           false,
		Verbose:         false,
		DryRun:          false,
		RecyclePoolSize: 100,
		PluginChanSize:  150,
		TickerLength:    10 * 60, // 10 minutes
		MaxMsgLoops:     4,
		MaxPackIdle:     idle,
		StartedAt:       time.Now(),
		Logger:          log.New(os.Stdout, "", log.Ldate|log.Lshortfile|log.Ltime),
	}
}
