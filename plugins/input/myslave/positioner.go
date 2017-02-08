package myslave

type positioner interface {

	// MarkAsProcessed tells the positioner that a certain binlog event
	// has been successfully processed and should be committed.
	MarkAsProcessed(file string, offset uint32) error

	// Committed returns the committed position.
	Committed() (file string, offset uint32, err error)

	// Flush will flush all processed binlog event position to storage.
	Flush() error
}
