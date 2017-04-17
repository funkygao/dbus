package engine

// Exchange is the packet tranport channel between plugins.
// Packet flows all through Exchange.
type Exchange interface {

	// InChan returns input channel from which Inputs can get reusable Packets.
	// The returned channel will be closed when engine stops.
	InChan() <-chan *Packet

	// Inject injects Packet into engine for consumers.
	Inject(pack *Packet)
}
