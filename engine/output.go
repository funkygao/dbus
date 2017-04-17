package engine

// Output is the output plugin.
type Output interface {
	Plugin

	// Run starts the main loop of the Output plugin.
	Run(r OutputRunner, h PluginHelper) (err error)
}

// OutputRunner is a helper for Output plugin to access some context data.
type OutputRunner interface {
	FilterOutputRunner

	// Output returns the underlying Output plugin.
	Output() Output

	// Ack notifies the packet's source Input plugin that it is processed successfully.
	Ack(*Packet) error
}
