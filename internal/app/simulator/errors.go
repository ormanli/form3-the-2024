package simulator

import "errors"

// ErrInvalidRequest represents an error indicating that a request is invalid.
var ErrInvalidRequest = errors.New("invalid request")

// ErrInvalidAmount represents an error indicating that the amount provided is invalid.
var ErrInvalidAmount = errors.New("invalid amount")
