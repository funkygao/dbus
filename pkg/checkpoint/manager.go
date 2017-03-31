package checkpoint

// Manager is an interface that manages checkpoint.
type Manager interface {

	// AllStates dumps all event state information.
	AllStates() ([]State, error)
}
