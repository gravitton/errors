package errors

import (
	"errors"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"maps"
	"runtime"
	"slices"
	"strings"
)

const maxFrames = 32

// Error is an immutable error with key-value fields, an optional cause, and a stack trace.
// The zero value has no underlying error and its message is <nil>.
type Error struct {
	err    error
	cause  error
	fields map[string]any
	stack  []uintptr
}

// New creates an Error with the given text and the current stack trace.
func New(text string) *Error {
	return wrap(errors.New(text))
}

// Sentinel creates a plain error without a stack trace, meant for package-level errors.
func Sentinel(text string) error {
	return errors.New(text)
}

// Newf formats an error with fmt.Errorf and wraps it like Wrap does.
func Newf(format string, args ...any) *Error {
	return wrap(fmt.Errorf(format, args...))
}

// Wrap returns err unchanged if it is an Error, or else converts it into one with the stack trace of
// the first Error in its Unwrap chain, which is closer to where the error was raised, or else the
// current one. A nil err returns nil.
func Wrap(err error) *Error {
	if err == nil {
		return nil
	}

	if e, ok := err.(*Error); ok {
		return e
	}

	return wrap(err)
}

func wrap(err error) *Error {
	return &Error{
		err:   err,
		stack: stackOf(err),
	}
}

// Error returns the message of the underlying error.
func (e *Error) Error() string {
	if e == nil || e.err == nil {
		return "<nil>"
	}

	return e.err.Error()
}

// Unwrap returns the underlying error, then the cause if there is one.
func (e *Error) Unwrap() []error {
	if e == nil || e.err == nil {
		return nil
	}
	if e.cause == nil {
		return []error{e.err}
	}

	return []error{e.err, e.cause}
}

// Is reports whether target is an Error whose underlying error matches this one's, so copies
// made by the With methods match the error they came from.
func (e *Error) Is(target error) bool {
	other, ok := target.(*Error)
	if !ok || e == nil || other == nil || other.err == nil {
		return false
	}

	return errors.Is(e.err, other.err)
}

// Fields returns a copy of the fields.
func (e *Error) Fields() map[string]any {
	if e == nil {
		return nil
	}

	return maps.Clone(e.fields)
}

// WithField returns a copy of the error with the field added. It panics on a nil error.
func (e *Error) WithField(key string, value any) *Error {
	return e.WithFields(map[string]any{key: value})
}

// WithFields returns a copy of the error with the fields added. It panics on a nil error.
func (e *Error) WithFields(fields map[string]any) *Error {
	derived := *e
	derived.fields = make(map[string]any, len(e.fields)+len(fields))
	maps.Copy(derived.fields, e.fields)
	maps.Copy(derived.fields, fields)

	return &derived
}

// WithCause returns a copy of the error with the cause attached, joined with the causes
// it already has into one flat collection. A nil or empty cause attaches nothing.
// It panics on a nil error.
func (e *Error) WithCause(cause error) *Error {
	causes := slices.Concat(members(e.cause), members(cause))

	derived := *e

	switch len(causes) {
	case 0:
	case 1:
		derived.cause = causes[0]
	default:
		derived.cause = Join(causes...)
	}

	return &derived
}

// StackTrace returns a copy of the program counters captured for the error,
// innermost call first.
func (e *Error) StackTrace() []uintptr {
	if e == nil {
		return nil
	}

	return slices.Clone(e.stack)
}

// Frames resolves the stack trace into call frames, innermost call first.
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

// Format prints the message for %s, %v and %q, the underlying error with fields,
// stack trace and cause for %+v, and Go syntax for %#v.
func (e *Error) Format(s fmt.State, verb rune) {
	switch {
	case verb == 'v' && s.Flag('#'):
		_, _ = io.WriteString(s, e.GoString())
	case verb == 'v' && s.Flag('+'):
		_, _ = io.WriteString(s, e.details())
	default:
		_, _ = fmt.Fprintf(s, fmt.FormatString(s, verb), e.Error())
	}
}

// GoString returns the error in Go syntax.
func (e *Error) GoString() string {
	if e == nil {
		return "(*errors.Error)(nil)"
	}

	var items []string

	if e.err != nil {
		items = append(items, fmt.Sprintf("err:%#v", e.err))
	}

	if e.cause != nil {
		items = append(items, fmt.Sprintf("cause:%#v", e.cause))
	}

	if len(e.fields) > 0 {
		items = append(items, fmt.Sprintf("fields:%#v", e.fields))
	}

	return "&errors.Error{" + strings.Join(items, ", ") + "}"
}

// LogValue returns the message, a group of the fields and the cause as a slog group,
// without the stack trace.
func (e *Error) LogValue() slog.Value {
	if e == nil {
		return slog.AnyValue(nil)
	}

	fields := make([]slog.Attr, 0, len(e.fields))
	for _, key := range slices.Sorted(maps.Keys(e.fields)) {
		fields = append(fields, slog.Any(key, e.fields[key]))
	}

	attrs := []slog.Attr{
		slog.String("msg", e.Error()),
		slog.GroupAttrs("fields", fields...),
	}

	if e.cause != nil {
		attrs = append(attrs, slog.Any("cause", e.cause))
	}

	return slog.GroupValue(attrs...)
}

func (e *Error) details() string {
	if e == nil {
		return "<nil>"
	}

	lines := []string{fmt.Sprintf("%+v", e.err)}

	for _, key := range slices.Sorted(maps.Keys(e.fields)) {
		lines = append(lines, indent(fmt.Sprintf("%s=%v", key, e.fields[key])))
	}

	for frame := range e.Frames() {
		lines = append(lines, indent(fmt.Sprintf("%s\n\t%s:%d", frame.Function, frame.File, frame.Line)))
	}

	for _, cause := range members(e.cause) {
		lines = append(lines, indent(fmt.Sprintf("caused by: %+v", cause)))
	}

	return strings.Join(lines, "\n")
}

func stackOf(err error) []uintptr {
	for ; err != nil; err = errors.Unwrap(err) {
		if inner, ok := err.(*Error); ok && inner != nil {
			return inner.stack
		}
	}

	return callers()
}

func members(err error) []error {
	if !exists(err) {
		return nil
	}

	if errs, ok := err.(*MultiError); ok {
		return errs.Unwrap()
	}

	return []error{err}
}

func exists(err error) bool {
	switch err := err.(type) {
	case nil:
		return false
	case *Error:
		return err != nil
	case *MultiError:
		return err != nil
	}

	return true
}

func indent(text string) string {
	return "\t" + strings.ReplaceAll(text, "\n", "\n\t")
}

func callers() []uintptr {
	var stack [maxFrames]uintptr
	n := runtime.Callers(5, stack[:])

	return slices.Clone(stack[:n])
}
