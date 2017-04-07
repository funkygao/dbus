package kafka

import (
	"errors"
)

var (
	ErrNotReady       = errors.New("not ready")
	ErrNotAllowed     = errors.New("not allowed")
	ErrStopping       = errors.New("stopping")
	ErrAlreadyClosed  = errors.New("consumer already closed")
	ErrConsumerBroken = errors.New("consumer conn broken")
)

var InvalidPartitionID = int32(-1)
