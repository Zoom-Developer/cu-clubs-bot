package errorz

import "errors"

var (
	InvalidCallbackData = errors.New("invalid callback data")
	InvalidState        = errors.New("invalid state")
	InvalidCode         = errors.New("invalid code")
	Forbidden           = errors.New("forbidden")
)
