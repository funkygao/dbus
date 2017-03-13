package engine

import (
	"fmt"

	conf "github.com/funkygao/jsconf"
)

type Plugin interface {
	Init(config *conf.Conf)
}

// If a Plugin implements Restarter, it will be called on restart.
// Return value determines whether restart it or run once.
type Restarter interface {
	CleanupForRestart() bool
}

// If a Plugin implements Pauser, it can pause/resume.
type Pauser interface {
	Pause() error
	Resume() error
}

// RegisterPlugin allows plugin to register itself to the engine.
func RegisterPlugin(name string, factory func() Plugin) {
	if _, present := availablePlugins[name]; present {
		panic(fmt.Sprintf("plugin[%s] cannot register twice", name))
	}

	availablePlugins[name] = factory
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
	name     string
	class    string
	disabled bool
}

func (this *pluginCommons) loadConfig(section *conf.Conf) {
	if this.name = section.String("name", ""); this.name == "" {
		panic(fmt.Sprintf("name is required"))
	}

	if this.class = section.String("class", ""); this.class == "" {
		this.class = this.name
	}
	this.disabled = section.Bool("disabled", false)
}
