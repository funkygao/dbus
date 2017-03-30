// Package checkpoint persists event log state information so that
// event log monitoring can resume from the last read event in the case of
// a restart or unexpected interruption.
package checkpoint

type Checkpoint interface {

	// Shutdown stops the checkpoint and flush all inflight states to be persisted.
	Shutdown() error

	// Commit queues the given event state information to be persisted.
	Commit(state State) error

	// LastPersistedState returns the last committed and persisted state.
	LastPersistedState(state State) error
}
