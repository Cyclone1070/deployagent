package todo

import (
	"errors"
	"fmt"
)

// -- Errors --

// StoreWriteError is returned when writing to a store fails.
type StoreWriteError struct {
	Cause error
}

func (e *StoreWriteError) Error() string { return fmt.Sprintf("failed to write store: %v", e.Cause) }
func (e *StoreWriteError) Unwrap() error { return e.Cause }

// -- Sentinels --

var (
	ErrInvalidStatus    = errors.New("invalid status")
	ErrEmptyDescription = errors.New("description cannot be empty")
)
