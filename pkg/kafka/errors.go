package kafka

import (
	"errors"
)

var (
	ErrNotReady   = errors.New("not ready")
	ErrNotAllowed = errors.New("not allowed")
)
