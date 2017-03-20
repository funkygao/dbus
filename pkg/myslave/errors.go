package myslave

import (
	"errors"
)

var (
	ErrInvalidRowFormat = errors.New("binlog must be ROW format")
)
