package engine

// Exchange is the packet tranport channel between plugins.
// Packet flows all through Exchange.
type Exchange interface {

	// InChan returns input channel from which Inputs can get reusable Packets.
	InChan() chan *Packet

	// Inject injects Packet into engine for consumers.
	Inject(pack *Packet)
}
