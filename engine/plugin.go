package engine

import (
	"fmt"

	conf "github.com/funkygao/jsconf"
	"github.com/funkygao/pretty"
)

type Plugin interface {
	Init(config *conf.Conf)
}

// If a Plugin implements CleanupForRestart, it will be called on restart
// Return value determines whether restart it or run once
type Restarting interface {
	CleanupForRestart() bool
}

func RegisterPlugin(name string, factory func() Plugin) {
	if _, present := availablePlugins[name]; present {
		panic(fmt.Sprintf("plugin[%s] cannot register twice", name))
	}

	availablePlugins[name] = factory
}

// A helper object to support delayed plugin creation
type PluginWrapper struct {
	name          string
	configCreator func() *conf.Conf
	pluginCreator func() Plugin
}

func (this *PluginWrapper) Create() (plugin Plugin) {
	plugin = this.pluginCreator()
	plugin.Init(this.configCreator())
	return
}

// common config directives for all plugins
type pluginCommons struct {
	name     string `json:"name"`
	class    string `json:"class"`
	ticker   int    `json:"ticker_interval"`
	disabled bool   `json:"disabled"`
	comment  string `json:"comment"`
}

func (this *pluginCommons) load(section *conf.Conf) {
	this.name = section.String("name", "")
	if this.name == "" {
		pretty.Printf("%# v\n", *section)
		panic(fmt.Sprintf("invalid plugin config: %v", *section))
	}

	this.class = section.String("class", "")
	if this.class == "" {
		this.class = this.name
	}
	this.comment = section.String("comment", "")
	this.ticker = section.Int("ticker_interval", Globals().TickerLength)
	this.disabled = section.Bool("disabled", false)
}
