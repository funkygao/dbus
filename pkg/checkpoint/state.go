package checkpoint

// State is an interface for all event state information.
type State interface {
	Marshal() []byte

	Unmarshal([]byte)

	String() string
}
