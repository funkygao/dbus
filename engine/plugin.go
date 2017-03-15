package engine

import (
	"fmt"

	conf "github.com/funkygao/jsconf"
)

// Plugin is the base interface for all plugins.
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
	Pause(InputRunner) error
	Resume(InputRunner) error
}

// Acker is a callback interface that is called when a packet
// is processed successfully.
type Acker interface {
	OnAck(*Packet) error
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

func (pw *pluginWrapper) Create() (plugin Plugin) {
	plugin = pw.pluginCreator()
	plugin.Init(pw.configCreator())
	return
}

// pluginCommons is the common config directives for all plugins.
type pluginCommons struct {
	name     string
	class    string
	disabled bool
}

func (pc *pluginCommons) loadConfig(section *conf.Conf) {
	if pc.name = section.String("name", ""); pc.name == "" {
		panic(fmt.Sprintf("name is required"))
	}

	if pc.class = section.String("class", ""); pc.class == "" {
		pc.class = pc.name
	}
	pc.disabled = section.Bool("disabled", false)
}
