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
// The zero value is an error with an empty message.
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

// Wrap converts err into an Error with the stack trace of the first Error in its tree, which
// is closer to where the error was raised, or else the current one. A nil err returns nil.
func Wrap(err error) *Error {
	if err == nil {
		return nil
	}

	return wrap(err)
}

func wrap(err error) *Error {
	if e, ok := err.(*Error); ok {
		return e
	}

	if inner, ok := errors.AsType[*Error](err); ok {
		return &Error{
			err:   err,
			stack: inner.stack,
		}
	}

	return &Error{
		err:   err,
		stack: callers(),
	}
}

// Error returns the message of the underlying error.
func (e *Error) Error() string {
	if e.err == nil {
		return ""
	}

	return e.err.Error()
}

// Unwrap returns the underlying error, then the cause if there is one.
func (e *Error) Unwrap() []error {
	if e.err == nil {
		return nil
	}

	if e.cause == nil {
		return []error{e.err}
	}

	return []error{e.err, e.cause}
}

// Fields returns a copy of the fields.
func (e *Error) Fields() map[string]any {
	return maps.Clone(e.fields)
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

// WithCause returns a copy of the error with the cause attached, joined with the cause
// it already has. A nil cause returns the error unchanged.
func (e *Error) WithCause(cause error) *Error {
	if cause == nil {
		return e
	}

	derived := *e
	derived.cause = cause

	if e.cause != nil {
		derived.cause = Join(e.cause, cause)
	}

	return &derived
}

// StackTrace returns a copy of the program counters captured for the error,
// innermost call first.
func (e *Error) StackTrace() []uintptr {
	return slices.Clone(e.stack)
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
	if e.err == nil {
		return "&errors.Error{}"
	}

	items := []string{fmt.Sprintf("err:%#v", e.err)}

	if e.cause != nil {
		items = append(items, fmt.Sprintf("cause:%#v", e.cause))
	}

	if len(e.fields) > 0 {
		items = append(items, fmt.Sprintf("fields:%#v", e.fields))
	}

	return "&errors.Error{" + strings.Join(items, ", ") + "}"
}

// LogValue returns the message, the fields and the cause as a slog group, without the
// stack trace.
func (e *Error) LogValue() slog.Value {
	attrs := []slog.Attr{slog.String("msg", e.Error())}

	for _, key := range slices.Sorted(maps.Keys(e.fields)) {
		attrs = append(attrs, slog.Any(key, e.fields[key]))
	}

	if e.cause != nil {
		attrs = append(attrs, slog.Any("cause", e.cause))
	}

	return slog.GroupValue(attrs...)
}

func (e *Error) details() string {
	message := ""
	if e.err != nil {
		message = fmt.Sprintf("%+v", e.err)
	}

	lines := []string{message}

	for _, key := range slices.Sorted(maps.Keys(e.fields)) {
		lines = append(lines, indent(fmt.Sprintf("%s=%v", key, e.fields[key])))
	}

	for frame := range e.frames() {
		lines = append(lines, indent(fmt.Sprintf("%s\n\t%s:%d", frame.Function, frame.File, frame.Line)))
	}

	if e.cause != nil {
		lines = append(lines, indent(fmt.Sprintf("caused by: %+v", e.cause)))
	}

	return strings.Join(lines, "\n")
}

func (e *Error) frames() iter.Seq[runtime.Frame] {
	return func(yield func(runtime.Frame) bool) {
		if len(e.stack) == 0 {
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

func indent(text string) string {
	return "\t" + strings.ReplaceAll(text, "\n", "\n\t")
}

func callers() []uintptr {
	var stack [maxFrames]uintptr
	n := runtime.Callers(4, stack[:])

	return slices.Clone(stack[:n])
}
