package engine

// Filter is the filter plugin.
type Filter interface {
	Plugin

	// Run starts the main loop of the Filter plugin.
	Run(r FilterRunner, h PluginHelper, stopper <-chan struct{}) (err error)
}

// FilterRunner is a helper for Filter plugin to access some context data.
type FilterRunner interface {
	FilterOutputRunner

	// Filter returns the underlying Filter plugin.
	Filter() Filter
}
