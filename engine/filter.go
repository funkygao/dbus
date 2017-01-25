package engine

type FilterRunner interface {
	FilterOutputRunner

	Filter() Filter
	Inject(pack *PipelinePack) bool
}

type Filter interface {
	Plugin

	Run(r FilterRunner, h PluginHelper) (err error)
}
