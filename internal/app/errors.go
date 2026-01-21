package app

import "errors"

// Sentinel errors for common application errors
var (
	ErrJobAlreadyExists   = errors.New("job already exists")
	ErrNotFound           = errors.New("not found")
	ErrInvalidArgument    = errors.New("invalid argument")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrDuplicateURL       = errors.New("a job with this URL already exists")
)
