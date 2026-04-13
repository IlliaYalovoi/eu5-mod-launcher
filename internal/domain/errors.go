package domain

import (
	"errors"
)

var (
	ErrInvalidID    = errors.New("invalid identifier")
	ErrTypeMismatch = errors.New("target type mismatch")
	ErrOutOfRange   = errors.New("value out of range")
	ErrEmptyValue   = errors.New("value must not be empty")
)
