package myslave

type positionManager interface {
	MarkAsProcessed(name string, position uint64)
}
