package intele

import "errors"

var (
	ErrTimeout           = errors.New("input timeout")
	ErrTooManyConcurrent = errors.New("too many concurrent requests")
)
