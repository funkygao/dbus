package cluster

import (
	"errors"
)

var (
	ErrResourceNotFound = errors.New("resource not found")
	ErrInvalidCallback  = errors.New("callback cannot be nil when resources present")
)
