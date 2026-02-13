package repository

import (
	"errors"
)

// Common repository errors
var (
	ErrNotFound     = errors.New("record not found")
	ErrDuplicate    = errors.New("duplicate record")
	ErrInvalidInput = errors.New("invalid input")
	ErrDatabase     = errors.New("database error")
)
