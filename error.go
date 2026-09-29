package errors

import (
	"errors"
	"fmt"
	"iter"
	"maps"
	"runtime"
	"slices"
)

const maxFrames = 32

// Error is an immutable error with key-value fields, causes, and a stack trace.
type Error struct {
	errs   []error
	fields map[string]any
	stack  []uintptr
}

// New creates an Error with the given text and the current stack trace.
func New(text string) *Error {
	return wrap(errors.New(text))
}

// Sentinel creates an Error without a stack trace, meant for package-level errors.
// The stack trace is captured where the error is derived or wrapped.
func Sentinel(text string) *Error {
	return &Error{
		errs: []error{errors.New(text)},
	}
}

// Newf formats an error with fmt.Errorf and wraps it like Wrap does.
func Newf(format string, args ...any) *Error {
	return wrap(fmt.Errorf(format, args...))
}

// Wrap converts err into an Error with the best stack trace available: the one of
// the first Error found through single-error wrappers, otherwise the current one.
// A nil err returns nil.
func Wrap(err error) *Error {
	if err == nil {
		return nil
	}

	return wrap(err)
}

func wrap(err error) *Error {
	derived := Error{
		errs: []error{err},
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
	return e.err().Error()
}

// Unwrap returns the underlying error followed by the causes.
func (e *Error) Unwrap() []error {
	return slices.Clip(e.errs)
}

// Is reports whether target is an Error whose underlying error is in the chain of
// this error's underlying error. Fields are not compared.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)

	return ok && errors.Is(e.err(), t.err())
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

// WithField returns a copy of the error with the field added. A copy of an error
// without a stack trace gets the current one.
func (e *Error) WithField(key string, value any) *Error {
	derived := e.derive()
	derived.fields = merged(e.fields, map[string]any{key: value})

	return &derived
}

// WithFields returns a copy of the error with the fields added. A copy of an error
// without a stack trace gets the current one.
func (e *Error) WithFields(fields map[string]any) *Error {
	derived := e.derive()
	derived.fields = merged(e.fields, fields)

	return &derived
}

// WithCause returns a copy of the error with the cause appended. A copy of an error
// without a stack trace gets the current one. A nil cause returns the error unchanged.
func (e *Error) WithCause(cause error) *Error {
	if cause == nil {
		return e
	}

	derived := e.derive()
	derived.errs = append(slices.Clip(e.errs), cause)

	return &derived
}

// StackTrace returns a copy of the program counters captured for the error,
// innermost call first.
func (e *Error) StackTrace() []uintptr {
	return slices.Clone(e.stack)
}

// Format prints the message for %s, %v and %q, the underlying error with fields,
// stack trace and causes for %+v, and Go syntax for %#v.
func (e *Error) Format(s fmt.State, verb rune) {
	format(e, s, verb)
}

// GoString returns the error in Go syntax.
func (e *Error) GoString() string {
	return fmt.Sprintf("&errors.Error{err:%#v, fields:%#v, causes:%s}", e.err(), e.fields, goSyntax(e.causes()))
}

func (e *Error) err() error {
	return e.errs[0]
}

func (e *Error) causes() []error {
	return e.errs[1:]
}

func (e *Error) derive() Error {
	derived := *e
	derived.stack = stackOf(e)

	return derived
}

func (e *Error) details() string {
	var children []string

	fields := e.Fields()
	for _, key := range slices.Sorted(maps.Keys(fields)) {
		children = append(children, fmt.Sprintf("%s=%v", key, fields[key]))
	}

	for frame := range frames(e.stack) {
		children = append(children, block(frame.Function, fmt.Sprintf("%s:%d", frame.File, frame.Line)))
	}

	for inner := range mainChain(e) {
		for _, cause := range inner.causes() {
			children = append(children, fmt.Sprintf("caused by: %+v", cause))
		}
	}

	return block(fmt.Sprintf("%+v", e.err()), children...)
}

func merged(base, fields map[string]any) map[string]any {
	merged := make(map[string]any, len(base)+len(fields))
	maps.Copy(merged, base)
	maps.Copy(merged, fields)

	return merged
}

func mainChain(err error) iter.Seq[*Error] {
	return func(yield func(*Error) bool) {
		for err != nil {
			switch e := err.(type) {
			case *Error:
				if !yield(e) {
					return
				}

				err = e.err()
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

	var stack [maxFrames]uintptr
	n := runtime.Callers(4, stack[:])

	return slices.Clone(stack[:n])
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
