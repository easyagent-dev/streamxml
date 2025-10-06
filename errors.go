package streamxml

import "errors"

// Error definitions for the parser
var (
	// ErrMaxDepthExceeded is returned when XML nesting exceeds the maximum allowed depth
	ErrMaxDepthExceeded = errors.New("maximum XML nesting depth exceeded")

	// ErrMaxBufferSizeExceeded is returned when the internal buffer exceeds the maximum allowed size
	ErrMaxBufferSizeExceeded = errors.New("maximum buffer size exceeded")

	// ErrInvalidConfiguration is returned when parser configuration is invalid
	ErrInvalidConfiguration = errors.New("invalid parser configuration")
)
