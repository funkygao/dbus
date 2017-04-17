// Package batcher provides retriable batch queue: all succeed or rollback for all.
// In kafka it is called RecordAccumulator.
package batcher

// Batcher is a batched queue with the sematics of all succeed and advance or any fails and retry.
type Batcher interface {

	// Close closes the batcher from R/W.
	Close()

	// Put writes an item to the batcher. If queue is full, it will block till
	// all inflight items marked success.
	Put(item interface{}) error

	// Get reads an item from the batcher.
	Get() (interface{}, error)

	// Succeed marks an item handling success.
	Succeed()

	// Fail marks an item handling failure.
	Fail() (rewind bool)
}
