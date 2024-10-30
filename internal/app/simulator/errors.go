package simulator

import "errors"

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrInvalidAmount  = errors.New("invalid amount")
)
