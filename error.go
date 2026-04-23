package errors

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
)

// DataError is an error enriched with structured key-value fields, an optional
// cause chain, and a captured stack trace. All methods that add data return a
// new copy; the original is never mutated.
type DataError struct {
	err   error
	data  map[string]any
	cause error
	stack []uintptr
}

// New creates a DataError from the given text and captures the current stack
// trace.
func New(text string) *DataError {
	return &DataError{
		err:   errors.New(text),
		stack: callers(1),
	}
}

// Newf creates a DataError from a formatted string and captures the current
// stack trace.
func Newf(format string, v ...any) *DataError {
	return &DataError{
		err:   fmt.Errorf(format, v...),
		stack: callers(1),
	}
}

// Wrap converts an error into a *DataError and captures the current stack
// trace. If err is nil, Wrap returns nil. If err is already a *DataError it is
// returned unchanged. Otherwise the error is wrapped directly.
//
// Warning: the returned *DataError nil is a typed nil pointer. When assigned to
// or returned as an error interface it will not equal nil. Prefer checking the
// error before passing it to Wrap rather than checking the result afterwards.
func Wrap(err error) *DataError {
	if err == nil {
		return nil
	}

	if dataErr, ok := err.(*DataError); ok {
		return dataErr
	}

	return &DataError{
		err:   err,
		stack: callers(1),
	}
}

// callers returns up to 32 program counters starting skip frames above the
// caller.
func callers(skip int) []uintptr {
	stack := make([]uintptr, 32)
	n := runtime.Callers(skip+2, stack)

	return stack[:n]
}

// Error returns the error message string.
func (e *DataError) Error() string {
	return e.err.Error()
}

// Fields returns the structured key-value data attached to this error.
func (e *DataError) Fields() map[string]any {
	return e.data
}

// StackTrace returns the program counters captured when the error was created.
func (e *DataError) StackTrace() []uintptr {
	return e.stack
}

// WithField returns a copy of the error with the given key-value field added.
// The original error is not modified.
func (e *DataError) WithField(key string, value any) *DataError {
	return e.WithFields(map[string]any{key: value})
}

// WithFields returns a copy of the error with the given fields merged in.
// The original error is not modified. Function values (including pointers to
// functions) are silently ignored because they are not safely comparable or
// serialisable.
func (e *DataError) WithFields(values map[string]any) *DataError {
	data := make(map[string]any, len(e.data)+len(values))
	for k, v := range e.data {
		data[k] = v
	}

	for k, v := range values {
		if t := reflect.TypeOf(v); t != nil {
			switch {
			case t.Kind() == reflect.Func, t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Func:
				continue
			}
		}

		data[k] = v
	}

	return &DataError{
		err:   e.err,
		data:  data,
		cause: e.cause,
		stack: e.stack,
	}
}

// WithCause returns a copy of the error with the given cause attached. The
// cause is returned by Unwrap, making it visible to errors.Is and errors.As.
func (e *DataError) WithCause(err error) *DataError {
	return &DataError{
		err:   e.err,
		data:  e.data,
		cause: err,
		stack: e.stack,
	}
}

// Unwrap returns the cause if one was set via WithCause; otherwise it returns
// the underlying error created by New, Newf, or Wrap.
func (e *DataError) Unwrap() error {
	if e.cause != nil {
		return e.cause
	}

	return e.err
}

// Is reports whether e matches target. Two *DataError values are considered
// equal when their messages match and every field present in target also
// appears in e with the same value. This allows errors.Is to find a sentinel
// DataError anywhere in a chain, optionally scoped by fields.
func (e *DataError) Is(target error) bool {
	var err *DataError
	ok := errors.As(target, &err)
	if !ok {
		return false
	}

	if e.Error() != err.Error() {
		return false
	}

	for k, v := range err.Fields() {
		if e.Fields()[k] != v {
			return false
		}
	}

	return true
}
