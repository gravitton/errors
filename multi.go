package errors

import (
	"fmt"
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

// Error returns the combined error message. An empty MultiError returns an
// empty string. A single-error MultiError returns that error's message
// unchanged. Otherwise a numbered summary is returned.
func (e *MultiError) Error() string {
	e.mutex.RLock()
	errs := e.errs
	e.mutex.RUnlock()

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

// GoString implements fmt.GoStringer for debugging output.
func (e *MultiError) GoString() string {
	e.mutex.RLock()
	errs := e.errs
	e.mutex.RUnlock()

	return fmt.Sprintf("%#v", errs)
}

// Unwrap returns the slice of collected errors, satisfying the Go 1.20+
// multi-error unwrap interface. It returns nil for a nil receiver.
func (e *MultiError) Unwrap() []error {
	if e == nil {
		return nil
	}

	e.mutex.RLock()
	defer e.mutex.RUnlock()

	return e.errs
}

// Add adds err to the collection. Nil errors are silently ignored. It is
// safe to call Add concurrently with other Add calls.
func (e *MultiError) Add(err error) {
	if err == nil {
		return
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.errs = append(e.errs, err)
}

// ErrorOrNil returns nil if no errors have been collected, or e itself
// otherwise. A nil receiver is treated as an empty collection and also returns
// nil.
func (e *MultiError) ErrorOrNil() error {
	if e == nil {
		return nil
	}

	e.mutex.RLock()
	n := len(e.errs)
	e.mutex.RUnlock()

	if n == 0 {
		return nil
	}

	return e
}
