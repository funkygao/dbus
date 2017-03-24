package cluster

// Controller is the contral brain of dbus cluster, which assigns resource tickets
// to participants.
//
// For dbus cluster to work, all dbus instances must share the same configurations,
// using zookeeper as configuration repo is highly suggetsted.
type Controller interface {
	Start() error

	Close() error

	//IsLeader() bool

	WaitForTicket(participant string) error

	RegisterResource(resource string) error

	RegisterParticipent(participant string) error
}
