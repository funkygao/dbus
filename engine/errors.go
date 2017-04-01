package engine

import "errors"

var (
	ErrInvalidParam = errors.New("invalid param")
	ErrQuitingSigal = errors.New("engine received quit signal")
)
