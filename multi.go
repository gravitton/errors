package errors

import (
	"fmt"
	"slices"
	"strings"
	"sync"
)

// MultiError collects errors into a single error. It is safe for concurrent use.
// The zero value is an empty collection.
type MultiError struct {
	errs  []error
	mutex sync.Mutex
}

// NewMulti returns an empty MultiError.
func NewMulti() *MultiError {
	return &MultiError{}
}

// Join returns a MultiError with the given errors, or nil when all of them are nil.
func Join(errs ...error) error {
	m := NewMulti()
	m.Add(errs...)

	return m.ErrorOrNil()
}

// Add adds the errors to the collection, skipping nils.
func (e *MultiError) Add(errs ...error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	for _, err := range errs {
		if err != nil {
			e.errs = append(e.errs, err)
		}
	}
}

// Error returns the messages of the collected errors joined by newlines.
func (e *MultiError) Error() string {
	return e.join(func(err error) string {
		return err.Error()
	})
}

// Unwrap returns the errors collected so far. Errors added later don't change the
// returned slice.
func (e *MultiError) Unwrap() []error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	return slices.Clip(e.errs)
}

// Len returns the number of collected errors.
func (e *MultiError) Len() int {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	return len(e.errs)
}

// ErrorOrNil returns nil when no errors were collected, otherwise the collection
// itself, which reflects errors added later.
func (e *MultiError) ErrorOrNil() error {
	if e.Len() == 0 {
		return nil
	}

	return e
}

// Format prints the message for %s, %v and %q, every error with %+v for %+v, and Go
// syntax for %#v.
func (e *MultiError) Format(s fmt.State, verb rune) {
	format(e, s, verb)
}

// GoString returns the collection in Go syntax.
func (e *MultiError) GoString() string {
	return fmt.Sprintf("&errors.MultiError{errs:%s}", goSyntax(e.Unwrap()))
}

func (e *MultiError) details() string {
	return e.join(func(err error) string {
		return fmt.Sprintf("%+v", err)
	})
}

func (e *MultiError) join(text func(error) string) string {
	errs := e.Unwrap()

	texts := make([]string, len(errs))
	for i, err := range errs {
		texts[i] = text(err)
	}

	return strings.Join(texts, "\n")
}
