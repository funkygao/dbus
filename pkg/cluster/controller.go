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

	// RegisterResource register a resource for an input plugin.
	RegisterResource(input string, resource string) error

	// RegisteredResources returns all the registered resource in the cluster.
	// The return map is in the form of {input: []resource}
	RegisteredResources() (map[string][]string, error)
}
