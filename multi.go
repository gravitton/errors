package errors

import (
	"fmt"
	"slices"
	"strings"
	"sync"
)

// MultiError collects errors into a single error. It is safe for concurrent use.
type MultiError struct {
	errs  []error
	mutex sync.RWMutex
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

// Error returns the single collected message unchanged, or a numbered, indented list.
func (e *MultiError) Error() string {
	return e.render(func(err error) string {
		return err.Error()
	})
}

// Unwrap returns a copy of the collected errors.
func (e *MultiError) Unwrap() []error {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	return slices.Clone(e.errs)
}

// Len returns the number of collected errors.
func (e *MultiError) Len() int {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	return len(e.errs)
}

// ErrorOrNil returns nil when no errors were collected, otherwise the collection.
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
	return fmt.Sprintf("&errors.MultiError{errs:%#v}", e.Unwrap())
}

func (e *MultiError) details() string {
	return e.render(func(err error) string {
		return fmt.Sprintf("%+v", err)
	})
}

func (e *MultiError) render(text func(error) string) string {
	errs := e.Unwrap()

	switch len(errs) {
	case 0:
		return "no errors"
	case 1:
		return text(errs[0])
	}

	lines := make([]string, len(errs))
	for i, err := range errs {
		lines[i] = indent(text(err))
	}

	return fmt.Sprintf("%d errors occurred:\n%s", len(errs), strings.Join(lines, "\n"))
}

func indent(text string) string {
	return "\t" + strings.ReplaceAll(text, "\n", "\n\t")
}
