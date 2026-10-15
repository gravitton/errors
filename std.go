package errors

import (
	"errors"
)

// ErrUnsupported is errors.ErrUnsupported.
var ErrUnsupported = errors.ErrUnsupported

// Unwrap calls errors.Unwrap, which returns nil for every Error: it follows only Unwrap() error.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// Is calls errors.Is.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As calls errors.As.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// AsType calls errors.AsType.
func AsType[E error](err error) (E, bool) {
	return errors.AsType[E](err)
}
