package todo

import (
	"errors"
	"fmt"
)

// -- Errors --

// StoreReadError is returned when reading from a store (e.g., todo) fails.
type StoreReadError struct {
	Cause error
}

func (e *StoreReadError) Error() string { return fmt.Sprintf("failed to read store: %v", e.Cause) }
func (e *StoreReadError) Unwrap() error { return e.Cause }

// StoreWriteError is returned when writing to a store fails.
type StoreWriteError struct {
	Cause error
}

func (e *StoreWriteError) Error() string { return fmt.Sprintf("failed to write store: %v", e.Cause) }
func (e *StoreWriteError) Unwrap() error { return e.Cause }

// -- Sentinels --

var (
	ErrInvalidStatus      = errors.New("invalid status")
	ErrEmptyDescription   = errors.New("description cannot be empty")
	ErrStoreNotConfigured = errors.New("todo store not configured")
)
