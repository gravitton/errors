package errors

import (
	"fmt"
	"slices"
	"strings"
	"sync"
)

// MultiError collects multiple errors into a single error value. It satisfies
// the Go 1.20+ multi-error unwrap interface (Unwrap() []error), so standard
// library functions such as errors.Is and errors.As traverse all collected
// errors. All methods are safe for concurrent use.
type MultiError struct {
	errs  []error
	mutex sync.RWMutex
}

// NewMulti returns a new, empty MultiError.
func NewMulti() *MultiError {
	return &MultiError{}
}

// Join returns a MultiError containing the given errors, skipping any nils.
// Returns nil if all arguments are nil.
func Join(errs ...error) error {
	m := &MultiError{}
	m.Add(errs...)

	return m.ErrorOrNil()
}

// Add adds the given errors to the collection. Nil errors are silently
// ignored. It is safe to call Add concurrently with other Add calls.
func (e *MultiError) Add(errs ...error) {
	if len(errs) == 0 {
		return
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	for _, err := range errs {
		if err != nil {
			e.errs = append(e.errs, err)
		}
	}
}

// Error returns the combined error message. An empty MultiError returns an
// empty string. A single-error MultiError returns that error's message
// unchanged. Otherwise a numbered summary is returned.
func (e *MultiError) Error() string {
	errs := e.Unwrap()

	switch len(errs) {
	case 0:
		return ""
	case 1:
		return errs[0].Error()
	default:
		msg := make([]string, len(errs))
		for i, err := range errs {
			msg[i] = err.Error()
		}

		return fmt.Sprintf("%d errors occurred:\n %s", len(errs), strings.Join(msg, "\n "))
	}
}

// Unwrap returns the collected errors, satisfying the Go 1.20+ multi-error
// unwrap interface. It returns nil for a nil receiver. The caller must not
// modify the returned slice.
func (e *MultiError) Unwrap() []error {
	if e == nil {
		return nil
	}

	e.mutex.RLock()
	defer e.mutex.RUnlock()

	return e.errs
}

// Errors returns a copy of the collected errors. Unlike [MultiError.Unwrap],
// the result is safe to modify.
func (e *MultiError) Errors() []error {
	return slices.Clone(e.Unwrap())
}

// Len returns the number of collected errors. A nil receiver reports zero.
func (e *MultiError) Len() int {
	if e == nil {
		return 0
	}

	e.mutex.RLock()
	defer e.mutex.RUnlock()

	return len(e.errs)
}

// ErrorOrNil returns nil if no errors have been collected, or e itself
// otherwise. A nil receiver is treated as an empty collection and also returns
// nil.
func (e *MultiError) ErrorOrNil() error {
	if e == nil || e.Len() == 0 {
		return nil
	}

	return e
}

// GoString implements fmt.GoStringer for debugging output.
func (e *MultiError) GoString() string {
	return fmt.Sprintf("%#v", e.Unwrap())
}
