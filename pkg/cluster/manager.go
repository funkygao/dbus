package cluster

// Manager is the interface that manages cluster state.
type Manager interface {

	// Open startup the manager.
	Open() error

	// Close closes the manager underlying connection.
	Close()

	// RegisterResource register a resource for an input plugin.
	RegisterResource(resource Resource) error

	// UnregisterResource removes a resource.
	UnregisterResource(resource Resource) error

	// RegisteredResources returns all the registered resource in the cluster.
	// The return map is in the form of {input: []resource}
	RegisteredResources() ([]Resource, error)

	// LiveParticipants returns currently online participants.
	LiveParticipants() ([]Participant, error)

	// Leader returns the controller leader participant.
	Leader() (Participant, error)

	// Rebalance triggers a new leader election across the cluster.
	Rebalance() error

	// TriggerUpgrade will notify all participants of binary upgrade.
	TriggerUpgrade() error

	// CurrentDecision returns current decision of the leader.
	CurrentDecision() Decision

	// CallParticipants calls each participants API specified by the query string.
	CallParticipants(method string, q string) error

	// Upgrade returns a channel to receive cluster upgrade events.
	Upgrade() <-chan struct{}
}
