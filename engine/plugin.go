package engine

import (
	"fmt"
	"net/http"
	"time"

	conf "github.com/funkygao/jsconf"
	"github.com/gorilla/mux"
)

type Plugin interface {
	Init(config *conf.Conf)
}

// If a Plugin implements CleanupForRestart, it will be called on restart.
// Return value determines whether restart it or run once.
type Restarting interface {
	CleanupForRestart() bool
}

// RegisterPlugin allows plugin to register itself to the engine.
func RegisterPlugin(name string, factory func() Plugin) {
	if _, present := availablePlugins[name]; present {
		panic(fmt.Sprintf("plugin[%s] cannot register twice", name))
	}

	availablePlugins[name] = factory
}

type PluginHelper interface {
	Engine() *Engine

	// TODO discard it
	PipelinePack(msgLoopCount int) *PipelinePack

	Project(name string) *ConfProject

	RegisterHttpApi(path string,
		handlerFunc func(http.ResponseWriter,
			*http.Request, map[string]interface{}) (interface{}, error)) *mux.Route
}

// pluginWrapper is a helper object to support delayed plugin creation.
type pluginWrapper struct {
	name string

	configCreator func() *conf.Conf
	pluginCreator func() Plugin
}

func (this *pluginWrapper) Create() (plugin Plugin) {
	plugin = this.pluginCreator()
	plugin.Init(this.configCreator())
	return
}

// pluginCommons is the common config directives for all plugins.
type pluginCommons struct {
	name     string        `json:"name"`
	class    string        `json:"class"`           // TODO
	ticker   time.Duration `json:"ticker_interval"` // TODO
	disabled bool          `json:"disabled"`
	comment  string        `json:"comment"`
}

func (this *pluginCommons) load(section *conf.Conf) {
	this.name = section.String("name", "")
	if this.name == "" {
		panic(fmt.Sprintf("name is required"))
	}

	this.class = section.String("class", "")
	if this.class == "" {
		this.class = this.name
	}
	this.comment = section.String("comment", "")
	this.ticker = section.Duration("ticker_interval", Globals().WatchdogTick)
	this.disabled = section.Bool("disabled", false)
}
