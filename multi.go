package errors

import (
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strconv"
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

// Join returns a MultiError with the given errors, or nil when Add skips all of them.
func Join(errs ...error) error {
	m := NewMulti()
	m.Add(errs...)

	return m.ErrorOrNil()
}

// Add adds the errors to the collection, skipping nils, nil Errors and nil MultiErrors.
// It panics on a nil collection.
func (e *MultiError) Add(errs ...error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	for _, err := range errs {
		if exists(err) {
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

// Unwrap returns a copy of the errors collected so far.
func (e *MultiError) Unwrap() []error {
	if e == nil {
		return nil
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	return slices.Clone(e.errs)
}

// Len returns the number of collected errors.
func (e *MultiError) Len() int {
	if e == nil {
		return 0
	}

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
	switch {
	case verb == 'v' && s.Flag('#'):
		_, _ = io.WriteString(s, e.GoString())
	case verb == 'v' && s.Flag('+'):
		_, _ = io.WriteString(s, e.details())
	default:
		_, _ = fmt.Fprintf(s, fmt.FormatString(s, verb), e.Error())
	}
}

// GoString returns the collection in Go syntax.
func (e *MultiError) GoString() string {
	if e == nil {
		return "(*errors.MultiError)(nil)"
	}

	errs := e.Unwrap()
	if len(errs) == 0 {
		return "&errors.MultiError{}"
	}

	items := make([]string, len(errs))
	for i, err := range errs {
		items[i] = fmt.Sprintf("%#v", err)
	}

	return "&errors.MultiError{errs:[]error{" + strings.Join(items, ", ") + "}}"
}

// LogValue returns the collected errors as a slog group keyed by their index.
func (e *MultiError) LogValue() slog.Value {
	if e == nil {
		return slog.AnyValue(nil)
	}

	errs := e.Unwrap()

	attrs := make([]slog.Attr, len(errs))
	for i, err := range errs {
		attrs[i] = slog.Any(strconv.Itoa(i), err)
	}

	return slog.GroupValue(attrs...)
}

func (e *MultiError) details() string {
	return e.join(func(err error) string {
		return fmt.Sprintf("%+v", err)
	})
}

func (e *MultiError) join(text func(error) string) string {
	if e == nil {
		return "<nil>"
	}

	errs := e.Unwrap()

	texts := make([]string, len(errs))
	for i, err := range errs {
		texts[i] = text(err)
	}

	return strings.Join(texts, "\n")
}
