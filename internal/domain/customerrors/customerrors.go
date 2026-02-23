package customerrors

import "errors"

//custom errors for better error handling and to avoid string comparisons in the codebase
var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrNotFound       = errors.New("resource not found")
	ErrAlreadyExists  = errors.New("resource already exists")
)
