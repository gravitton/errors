package errors

import (
	"errors"
)

// Unwrap only redirects to errors.Unwrap
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// Is only redirects to errors.Is
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As only redirects to errors.As
func As(err error, target any) bool {
	return errors.As(err, target)
}
