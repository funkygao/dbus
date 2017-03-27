package engine

import "errors"

var (
	ErrInvalidParam    = errors.New("invalid param")
	ErrClusterDisabled = errors.New("cluster disabled")
	ErrQuitingSigal    = errors.New("engine received quit signal")
	ErrDupResource     = errors.New("dup resource found")
)
