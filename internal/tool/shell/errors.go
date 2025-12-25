package shell

import (
	"errors"
	"fmt"
)

// -- Shell Tool Errors --

// EnvFileReadError is returned when reading an env file fails.
type EnvFileReadError struct {
	Path  string
	Cause error
}

func (e *EnvFileReadError) Error() string {
	return fmt.Sprintf("failed to read env file %s: %v", e.Path, e.Cause)
}
func (e *EnvFileReadError) Unwrap() error { return e.Cause }

// -- Shell Tool Sentinels --

var (
	ErrCommandRequired = errors.New("command cannot be empty")
	ErrInvalidTimeout  = errors.New("timeout_seconds cannot be negative")
	ErrEnvFileParse    = errors.New("invalid line in env file")
)
