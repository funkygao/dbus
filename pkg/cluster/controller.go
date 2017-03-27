// Package cluster provides helix-alike leader/standby model.
//
// It rebalances resources to participants and watch for cluster changes.
package cluster

// Controller is the contral brain of dbus cluster, which assigns resource tickets
// to participants.
//
// Controller works by assigning resources to participants according to the registry.
//
// For dbus cluster to work, all dbus instances must share the same configuration,
// using zookeeper as configuration repo is highly suggetsted.
type Controller interface {

	// Start startup the controller and start election across the cluster.
	Start() error

	// Close closes the controller underlying connection and do all the related cleanup.
	Close() error

	// IsLeader returns whether current participant is leader of the cluster.
	IsLeader() bool

	// RegisterResources notifies the controller all configured resources.
	RegisterResources([]string) error

	// DecodeResource decodes the resource into original format.
	//
	// IMPORTANT RegisterResources might encode the resource, so client need to decode resource.
	DecodeResource(encodedResource string) (string, error)
}
