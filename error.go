package errors

import (
	"errors"
	"fmt"
	"io"
	"iter"
	"maps"
	"reflect"
	"runtime"
	"slices"
	"strings"
)

// Error is an error enriched with structured key-value fields, an optional
// cause chain, and a captured stack trace. All methods that add data return a
// new copy; the original is never mutated.
type Error struct {
	err   error
	data  map[string]any
	cause error
	stack []uintptr
}

// New creates an Error from the given text and captures the current stack trace.
func New(text string) *Error {
	return &Error{
		err:   errors.New(text),
		stack: callers(1),
	}
}

// Newf creates an Error from a formatted string and captures the current stack trace.
func Newf(format string, v ...any) *Error {
	return &Error{
		err:   fmt.Errorf(format, v...),
		stack: callers(1),
	}
}

// Wrap converts an error into an *Error and captures the current stack trace.
// If err is nil, Wrap returns nil. If err is already an *Error it is returned
// unchanged. Otherwise the error is wrapped directly.
//
// Warning: the returned *Error nil is a typed nil pointer. When assigned to
// or returned as an error interface it will not equal nil. Prefer checking the
// error before passing it to Wrap rather than checking the result afterwards.
func Wrap(err error) *Error {
	if err == nil {
		return nil
	}

	if dataErr, ok := err.(*Error); ok {
		return dataErr
	}

	return &Error{
		err:   err,
		stack: callers(1),
	}
}

// Error returns the error message string. A nil *Error reports "<nil>".
func (e *Error) Error() string {
	if e == nil || e.err == nil {
		return "<nil>"
	}

	return e.err.Error()
}

// Fields returns a copy of the structured key-value data attached to this error.
// Mutating the result does not affect the error.
func (e *Error) Fields() map[string]any {
	if e == nil {
		return nil
	}

	return maps.Clone(e.data)
}

// WithField returns a copy of the error with the given key-value field added.
// The original error is not modified.
func (e *Error) WithField(key string, value any) *Error {
	return e.WithFields(map[string]any{key: value})
}

// WithFields returns a copy of the error with the given fields merged in.
// The original error is not modified. A nil *Error returns nil, so chaining after Wrap(nil) does not panic.
func (e *Error) WithFields(values map[string]any) *Error {
	if e == nil {
		return nil
	}

	data := make(map[string]any, len(e.data)+len(values))
	maps.Copy(data, e.data)
	maps.Copy(data, values)

	return &Error{
		err:   e.err,
		data:  data,
		cause: e.cause,
		stack: e.stack,
	}
}

// WithCause returns a copy of the error with the given cause attached.
// The cause is returned by Unwrap, making it visible to errors.Is and errors.As. A nil *Error returns nil.
func (e *Error) WithCause(err error) *Error {
	if e == nil {
		return nil
	}

	return &Error{
		err:   e.err,
		data:  e.data,
		cause: err,
		stack: e.stack,
	}
}

// Unwrap returns the cause if one was set via WithCause; otherwise it returns
// the underlying error created by New, Newf, or Wrap.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}

	if e.cause != nil {
		return e.cause
	}

	return e.err
}

// Is reports whether e matches target. Two *Error values are considered equal
// when their messages match and every field present in target also appears in
// e with the same value. This allows errors.Is to find a sentinel Error
// anywhere in a chain, optionally scoped by fields.
//
// Field values of different types never match. Uncomparable values are
// compared with reflect.DeepEqual, so function fields only match when both are nil.
func (e *Error) Is(target error) bool {
	var err *Error
	if e == nil || !errors.As(target, &err) || err == nil {
		return false
	}

	if e.Error() != err.Error() {
		return false
	}

	for k, v := range err.data {
		if !equal(e.data[k], v) {
			return false
		}
	}

	return true
}

// Format implements fmt.Formatter. The %s, %q and %v verbs print the message
// alone; %+v additionally prints the fields, the cause chain and the stack
// trace.
func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			io.WriteString(s, e.details())

			return
		}

		io.WriteString(s, e.Error())
	case 's':
		io.WriteString(s, e.Error())
	case 'q':
		fmt.Fprintf(s, "%q", e.Error())
	default:
		fmt.Fprintf(s, "%%!%c(*errors.Error=%s)", verb, e.Error())
	}
}

// StackTrace returns the program counters captured when the error was created.
// See [Error.Frames] for the resolved call frames.
func (e *Error) StackTrace() []uintptr {
	if e == nil {
		return nil
	}

	return e.stack
}

// Frames resolves the captured stack trace into call frames, outermost call first.
// The sequence is empty when no stack was captured.
func (e *Error) Frames() iter.Seq[runtime.Frame] {
	return func(yield func(runtime.Frame) bool) {
		if e == nil || len(e.stack) == 0 {
			return
		}

		frames := runtime.CallersFrames(e.stack)
		for {
			frame, more := frames.Next()
			if !yield(frame) || !more {
				return
			}
		}
	}
}

// details renders the message together with the fields, the cause and the
// stack trace, as printed by the %+v verb.
func (e *Error) details() string {
	b := &strings.Builder{}
	b.WriteString(e.Error())

	if e == nil {
		return b.String()
	}

	for _, k := range slices.Sorted(maps.Keys(e.data)) {
		fmt.Fprintf(b, "\n\t%s=%v", k, e.data[k])
	}

	if e.cause != nil {
		fmt.Fprintf(b, "\ncaused by: %v", e.cause)
	}

	for frame := range e.Frames() {
		fmt.Fprintf(b, "\n\t%s\n\t\t%s:%d", frame.Function, frame.File, frame.Line)
	}

	return b.String()
}

// callers returns up to 32 program counters starting skip frames above the caller.
func callers(skip int) []uintptr {
	stack := make([]uintptr, 32)
	n := runtime.Callers(skip+2, stack)

	return stack[:n]
}

// equal compares two field values without panicking on uncomparable types.
func equal(a, b any) bool {
	t := reflect.TypeOf(a)
	if t != reflect.TypeOf(b) {
		return false
	}

	if t == nil || t.Comparable() {
		return a == b
	}

	return reflect.DeepEqual(a, b)
}
