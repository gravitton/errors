package errors

import (
	"errors"
)

// ErrUnsupported is re-exported from the standard library so that callers can
// swap the import without losing it.
var ErrUnsupported = errors.ErrUnsupported

// Unwrap delegates to errors.Unwrap.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// Is delegates to errors.Is.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As delegates to errors.As.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// AsType delegates to errors.AsType.
func AsType[E error](err error) (E, bool) {
	return errors.AsType[E](err)
}
