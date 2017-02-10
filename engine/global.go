package engine

import (
	"fmt"
	"os"
	"regexp"
	"sync"
	"syscall"
	"time"

	conf "github.com/funkygao/jsconf"
)

const (
	RELOAD  = "reload"
	STOP    = "stop"
	SIGUSR1 = "user1"
	SIGUSR2 = "user2"
)

var (
	availablePlugins = make(map[string]func() Plugin) // name:factory
	pluginTypeRegex  = regexp.MustCompile("^.*(Filter|Input|Output)$")

	Globals func() *GlobalConfig
)

// GlobalConfig is the struct for holding global pipeline config values.
type GlobalConfig struct {
	*conf.Conf

	StartedAt       time.Time
	Stopping        bool
	Debug           bool
	DryRun          bool
	RecyclePoolSize int
	PluginChanSize  int

	WatchdogTick time.Duration

	MaxMsgLoops int
	MaxPackIdle time.Duration

	// registry is used to hold the global object shared between plugins.
	registry map[string]interface{}
	regMu    sync.RWMutex

	sigChan chan os.Signal
}

func (this *GlobalConfig) Shutdown() {
	this.Kill(syscall.SIGINT)
}

func (this *GlobalConfig) Kill(sig os.Signal) {
	this.sigChan <- sig
}

func (this *GlobalConfig) Register(k string, v interface{}) {
	this.regMu.Lock()
	defer this.regMu.Unlock()

	if _, present := this.registry[k]; present {
		panic(fmt.Sprintf("dup register: %s", k))
	}

	this.registry[k] = v
}

func (this *GlobalConfig) Registered(k string) interface{} {
	this.regMu.RLock()
	defer this.regMu.RUnlock()

	return this.registry[k]
}

func DefaultGlobals() *GlobalConfig {
	idle, _ := time.ParseDuration("2m")
	return &GlobalConfig{
		Debug:           false,
		DryRun:          false,
		RecyclePoolSize: 100,
		PluginChanSize:  150,
		WatchdogTick:    time.Minute * 10,
		MaxMsgLoops:     4,
		MaxPackIdle:     idle,
		StartedAt:       time.Now(),
		registry:        map[string]interface{}{},
	}
}
