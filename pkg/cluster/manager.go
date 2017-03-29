package cluster

// Manager is the interface that manages cluster state.
type Manager interface {

	// Open startup the manager.
	Open() error

	// Close closes the manager underlying connection.
	Close()

	// RegisterResource register a resource for an input plugin.
	RegisterResource(input string, resource string) error

	// RegisteredResources returns all the registered resource in the cluster.
	// The return map is in the form of {input: []resource}
	RegisteredResources() (map[string][]string, error)
}
