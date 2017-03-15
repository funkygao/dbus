package engine

type Acker interface {
	OnAck(*Packet) error
}
