// Package cluster provides helix-alike leader/standby model and layered resource scheduler.
//
// A cluster is composed of 1) participants 2) predefined resources(shard-able),
//
// One of the participant is the controller participant which schedules resources to
// participants by RPC, whence each participant schedules their own resources to
// local Input plugins.
//
// It rebalances resources to participants and watch for cluster changes.
//
// The cluster changes might be:
// participants come and go, resource added and deleted, leader change.
//
// The resource scheduler is layered dispatch mechanism:
//
//
//                    +------------+
//                    | controller |
//                    +------------+
//                         | participant:[]resource
//        +----------------------------------+
//        | RPC            | RPC             | RPC
//  +-------------+  +-------------+  +-------------+
//  | participant |  | participant |  | participant |
//  | RPC handler |  | RPC handler |  | RPC handler |
//  +-------------+  +-------------+  +-------------+
//                             |
//                     +-------------------+
//                     | []resource        | []resource
//                   +-------------+  +-------------+
//                   | InputPlugin |  | InputPlugin |
//                   +-------------+  +-------------+
//
package cluster

// RebalanceCallback connects cluster with its caller when leader decides to rebalance.
type RebalanceCallback func(epoch int, decision Decision)

// Controller is the contral brain of the cluster, which assigns resource tickets
// to participants.
//
// Controller works by assigning resources to participants if necessary.
//
// For dbus cluster to work, all dbus instances must share the same configuration,
// using zookeeper as configuration repo is highly suggetsted.
type Controller interface {

	// Start startup the controller and start election across the cluster.
	Start() error

	// Stop closes the controller underlying connection and do all the related cleanup.
	Stop() error

	// Upgrade returns a channel to receive cluster upgrade events.
	Upgrade() <-chan struct{}
}
