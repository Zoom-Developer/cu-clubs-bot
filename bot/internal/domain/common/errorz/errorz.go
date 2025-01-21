package errorz

import "errors"

var (
	ErrInvalidCallbackData = errors.New("invalid callback data")
	ErrInvalidState        = errors.New("invalid state")
	ErrInvalidCode         = errors.New("invalid code")
	ErrForbidden           = errors.New("forbidden")
)
