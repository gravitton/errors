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

// Error returns the combined error message. A nil or empty MultiError returns
// an empty string. A single-error MultiError returns that error's message
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

// Unwrap returns a copy of the collected errors, satisfying the Go 1.20+
// multi-error unwrap interface. It returns nil for a nil receiver or an empty
// collection. The result is safe to modify.
func (e *MultiError) Unwrap() []error {
	if e == nil {
		return nil
	}

	e.mutex.RLock()
	defer e.mutex.RUnlock()

	return slices.Clone(e.errs)
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

// Format implements fmt.Formatter. The combined message is printed like a
// plain string, so %s, %q, %x and %v honour width, precision and flags. %+v
// formats every collected error with %+v as well, so the fields, causes and
// stack traces they carry are printed, and %#v prints the collection in Go
// syntax.
func (e *MultiError) Format(s fmt.State, verb rune) {
	format(e, s, verb)
}

// GoString implements fmt.GoStringer for debugging output.
func (e *MultiError) GoString() string {
	return fmt.Sprintf("%#v", e.Unwrap())
}

// details renders the collected errors with their details, as printed by the
// %+v verb. It mirrors Error: empty for no errors, the sole error unchanged,
// and otherwise a numbered summary with every error indented.
func (e *MultiError) details() string {
	errs := e.Unwrap()

	switch len(errs) {
	case 0:
		return ""
	case 1:
		return fmt.Sprintf("%+v", errs[0])
	default:
		msg := make([]string, len(errs))
		for i, err := range errs {
			msg[i] = indent(fmt.Sprintf("%+v", err))
		}

		return fmt.Sprintf("%d errors occurred:\n%s", len(errs), strings.Join(msg, "\n"))
	}
}
