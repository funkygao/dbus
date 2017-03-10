package engine

import (
	"github.com/gorilla/mux"
)

// PluginHelper is a helper for plugins to access partial functions of the singleton Engine.
type PluginHelper interface {

	// ClonePacket is used for plugin Filter to generate new Packet.
	ClonePacket(*Packet) *Packet

	// RegisterAPI allows plugins to register handlers on the global API server.
	RegisterAPI(path string, handlerFunc APIHandler) *mux.Route
}
