package cluster

// Controller is the contral brain of dbus cluster, which assigns resource tickets
// to participants.
//
// Controller works by assigning resources to participants according to the registry.
//
// For dbus cluster to work, all dbus instances must share the same configurations,
// using zookeeper as configuration repo is highly suggetsted.
type Controller interface {

	// Start startup the controller and start election across the cluster.
	Start() error

	// Close closes the controller underlying connection and do all the related cleanup.
	Close() error

	// IsLeader returns whether current participant is leader of the cluster.
	IsLeader() bool

	// RegisterParticipent registers a participant instance acorss the cluster.
	RegisterParticipent(participant string, weight int) error
}
