package customerrors

import "errors"

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrNotFound       = errors.New("resource not found")
	ErrAlreadyExists  = errors.New("resource already exists")
)
