package errors

import (
	"errors"
)

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
