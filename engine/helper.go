package engine

// PluginHelper is a helper for plugins to access partial functions of the singleton Engine.
type PluginHelper interface {
	// Project returns a Project by name.
	Project(name string) *Project

	// NewPacket is used for plugin Filter to generate new Packet.
	NewPacket() *Packet
}
