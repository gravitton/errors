package errors

import (
	"errors"
	"fmt"
	"iter"
	"maps"
	"runtime"
	"slices"
	"strings"
)

const maxFrames = 32

// Error is an immutable error with key-value fields, causes, and a stack trace.
type Error struct {
	err    error
	fields map[string]any
	causes []error
	stack  []uintptr
}

// New creates an Error with the given text and the current stack trace.
func New(text string) *Error {
	return wrap(errors.New(text))
}

// Sentinel creates an Error without a stack trace, meant for package-level errors.
// The stack trace is captured where it is passed to Wrap or wrapped by Newf.
func Sentinel(text string) *Error {
	return &Error{
		err: errors.New(text),
	}
}

// Newf creates an Error from fmt.Errorf, like Wrap does.
func Newf(format string, args ...any) *Error {
	return wrap(fmt.Errorf(format, args...))
}

// Wrap converts err into an Error with the best stack trace available: the one of
// the first Error in the main chain that has one, otherwise the current one.
// A nil err returns nil.
func Wrap(err error) *Error {
	if err == nil {
		return nil
	}

	return wrap(err)
}

func wrap(err error) *Error {
	derived := Error{
		err: err,
	}

	if e, ok := err.(*Error); ok {
		if e.stack != nil {
			return e
		}

		derived = *e
	}

	derived.stack = stackOf(err)

	return &derived
}

// Error returns the message of the underlying error.
func (e *Error) Error() string {
	return e.err.Error()
}

// Unwrap returns the underlying error followed by the causes.
func (e *Error) Unwrap() []error {
	return append([]error{e.err}, e.causes...)
}

// Is reports whether target is an Error whose underlying error is in the chain of
// this error's underlying error. Fields are not compared.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)

	return ok && errors.Is(e.err, t.err)
}

// Fields returns the fields of every Error in the main chain, merged. The outermost
// value of a key wins.
func (e *Error) Fields() map[string]any {
	fields := make(map[string]any)

	for inner := range mainChain(e) {
		for key, value := range inner.fields {
			if _, ok := fields[key]; !ok {
				fields[key] = value
			}
		}
	}

	return fields
}

// WithField returns a copy of the error with the field added.
func (e *Error) WithField(key string, value any) *Error {
	return e.WithFields(map[string]any{key: value})
}

// WithFields returns a copy of the error with the fields added.
func (e *Error) WithFields(fields map[string]any) *Error {
	derived := *e
	derived.fields = make(map[string]any, len(e.fields)+len(fields))
	maps.Copy(derived.fields, e.fields)
	maps.Copy(derived.fields, fields)

	return &derived
}

// WithCause returns a copy of the error with the cause appended. A nil cause
// returns the error unchanged.
func (e *Error) WithCause(cause error) *Error {
	if cause == nil {
		return e
	}

	derived := *e
	derived.causes = append(slices.Clip(e.causes), cause)

	return &derived
}

// StackTrace returns a copy of the program counters captured for the error,
// innermost call first.
func (e *Error) StackTrace() []uintptr {
	return slices.Clone(e.stack)
}

// Format prints the message for %s, %v and %q, the message with fields, stack
// trace and causes for %+v, and Go syntax for %#v.
func (e *Error) Format(s fmt.State, verb rune) {
	format(e, s, verb)
}

// GoString returns the error in Go syntax.
func (e *Error) GoString() string {
	return fmt.Sprintf("&errors.Error{err:%#v, fields:%#v, causes:%#v}", e.err, e.fields, e.causes)
}

func (e *Error) details() string {
	b := &strings.Builder{}
	b.WriteString(e.err.Error())

	fields := e.Fields()
	for _, key := range slices.Sorted(maps.Keys(fields)) {
		fmt.Fprintf(b, "\n\t%s=%v", key, fields[key])
	}

	for frame := range frames(e.stack) {
		fmt.Fprintf(b, "\n\t%s\n\t\t%s:%d", frame.Function, frame.File, frame.Line)
	}

	for inner := range mainChain(e) {
		for _, cause := range inner.causes {
			fmt.Fprintf(b, "\ncaused by: %+v", cause)
		}
	}

	return b.String()
}

func mainChain(err error) iter.Seq[*Error] {
	return func(yield func(*Error) bool) {
		for err != nil {
			switch e := err.(type) {
			case *Error:
				if !yield(e) {
					return
				}

				err = e.err
			case interface{ Unwrap() error }:
				err = e.Unwrap()
			default:
				return
			}
		}
	}
}

func stackOf(err error) []uintptr {
	for e := range mainChain(err) {
		if e.stack != nil {
			return e.stack
		}
	}

	stack := make([]uintptr, maxFrames)
	n := runtime.Callers(4, stack)

	return stack[:n]
}

func frames(stack []uintptr) iter.Seq[runtime.Frame] {
	return func(yield func(runtime.Frame) bool) {
		if len(stack) == 0 {
			return
		}

		frames := runtime.CallersFrames(stack)
		for {
			frame, more := frames.Next()
			if !yield(frame) || !more {
				return
			}
		}
	}
}
