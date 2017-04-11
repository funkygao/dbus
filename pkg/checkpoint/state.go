package checkpoint

// State is an interface for all event state information.
type State interface {

	// Marshal serialize the state data.
	Marshal() []byte

	// Unmarshal unserialize the state from byte array data.
	Unmarshal([]byte)

	// String returns the string form of the state position.
	String() string

	// DSN returns the data source of the state.
	DSN() string

	// Scheme returns the type of statte.
	Scheme() string

	// Name returns name of the state.
	Name() string
}
