package cluster

import (
	"errors"
)

var (
	ErrResourceNotFound   = errors.New("resource not found")
	ErrResourceDuplicated = errors.New("dup resource not allowed")
)
