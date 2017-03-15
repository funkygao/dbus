package engine

import "errors"

var (
	ErrInvalidParam = errors.New("invalid param")
	ErrNotFound     = errors.New("not found")
)
