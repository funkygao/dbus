// Package checkpoint persists event log state information so that
// event log monitoring can resume from the last read event in the case of
// a restart or unexpected interruption.
package checkpoint

type Checkpoint interface {

	// Start startup the checkpoint and be ready to receive state information.
	Start() error

	// Stop stops the checkpoint and flush all inflight states to be persisted.
	Stop() error

	// Commit queues the given event state information to be persisted.
	Commit(state State) error

	// State returns the last committed state.
	State() (state State, err error)
}
