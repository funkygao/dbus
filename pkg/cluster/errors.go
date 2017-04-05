package cluster

import (
	"errors"
)

var (
	ErrNoLeader = errors.New("no leader found")
)
