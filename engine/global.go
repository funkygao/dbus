package engine

import (
	"fmt"
	"os"
	"regexp"
	"sync"
	"syscall"
	"time"

	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gafka/zk"
	conf "github.com/funkygao/jsconf"
)

const (
	RELOAD  = "reload"
	STOP    = "stop"
	SIGUSR1 = "user1"
	SIGUSR2 = "user2"

	DbusConfZnode = "/dbus/conf"
)

var (
	availablePlugins = make(map[string]func() Plugin) // name:factory
	pluginTypeRegex  = regexp.MustCompile("^.*(Filter|Input|Output)$")

	Globals func() *GlobalConfig
)

// GlobalConfig is the struct for holding global pipeline config values.
type GlobalConfig struct {
	*conf.Conf

	StartedAt   time.Time
	Stopping    bool
	Debug       bool
	RouterTrack bool

	InputRecyclePoolSize  int
	FilterRecyclePoolSize int
	HubChanSize           int
	PluginChanSize        int

	WatchdogTick time.Duration

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

func (this *GlobalConfig) GetOrRegisterZkzone(zone string) *zk.ZkZone {
	this.regMu.Lock()
	defer this.regMu.Unlock()

	key := fmt.Sprintf("zkzone.%s", zone)
	if _, present := this.registry[key]; !present {
		zkzone := zk.NewZkZone(zk.DefaultConfig(zone, ctx.ZoneZkAddrs(zone)))
		this.registry[key] = zkzone
	}

	return this.registry[key].(*zk.ZkZone)
}

func DefaultGlobals() *GlobalConfig {
	return &GlobalConfig{
		Debug:                 false,
		InputRecyclePoolSize:  100,
		FilterRecyclePoolSize: 100,
		HubChanSize:           200,
		PluginChanSize:        150,
		RouterTrack:           true,
		WatchdogTick:          time.Minute * 2,
		StartedAt:             time.Now(),
		registry:              map[string]interface{}{},
	}
}
