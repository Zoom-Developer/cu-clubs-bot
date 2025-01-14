package errorz

import "errors"

var (
	InvalidCallbackData = errors.New("invalid callback data")
	InvalidState        = errors.New("invalid state")
	Forbidden           = errors.New("forbidden")
)
