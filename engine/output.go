package engine

type Output interface {
	Plugin

	Run(r OutputRunner, h PluginHelper) (err error)
}

type OutputRunner interface {
	FilterOutputRunner

	Output() Output
}
