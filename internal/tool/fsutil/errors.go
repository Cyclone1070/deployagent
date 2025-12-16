package fsutil

import "errors"

// ErrInvalidOffset is returned when offset parameter is negative.
var ErrInvalidOffset = errors.New("offset must be >= 0")
