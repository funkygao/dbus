package engine

type OutputRunner interface {
	FilterOutputRunner

	Output() Output
}

type Output interface {
	Plugin

	Run(r OutputRunner, h PluginHelper) (err error)
}
