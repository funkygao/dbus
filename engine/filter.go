package engine

type Filter interface {
	Plugin

	Run(r FilterRunner, h PluginHelper) (err error)
}

type FilterRunner interface {
	FilterOutputRunner

	Filter() Filter
	Inject(pack *PipelinePack) bool
}
