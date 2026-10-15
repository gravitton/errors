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

// Error is an error with key-value fields, an optional cause and a stack trace. The zero value
// has no underlying error, cause or stack trace, and its message is <nil>.
type Error struct {
	err    error
	cause  error
	fields map[string]any
	stack  []uintptr
}

// New returns an Error with the given text and the current stack trace.
func New(text string) *Error {
	return wrap(errors.New(text))
}

// Sentinel returns a plain error without a stack trace, for package-level variables. Wrap
// derives an Error from it.
func Sentinel(text string) error {
	return errors.New(text)
}

// Newf returns Wrap(fmt.Errorf(format, args...)).
func Newf(format string, args ...any) *Error {
	return wrap(fmt.Errorf(format, args...))
}

// Wrap returns err itself if it is an Error, else a new Error around it with the stack trace of
// the first Error in its Unwrap() error chain, or the current one if there is none. A nil err
// returns nil.
func Wrap(err error) *Error {
	if !exists(err) {
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

// Is reports whether target is an Error whose underlying error errors.Is finds in the error's
// underlying error or cause. A target without an underlying error matches only itself.
func (e *Error) Is(target error) bool {
	other, ok := target.(*Error)
	if !ok || other == nil || other.err == nil {
		return false
	}

	return errors.Is(e, other.err)
}

// Fields returns the error's own fields, not those of its underlying error or cause.
func (e *Error) Fields() map[string]any {
	if e == nil {
		return nil
	}

	return maps.Clone(e.fields)
}

// WithField returns the error with the field added, replacing one with the same key. It panics
// for a nil receiver.
func (e *Error) WithField(key string, value any) *Error {
	return e.WithFields(map[string]any{key: value})
}

// WithFields returns the error with the fields added, replacing those with the same keys. It
// panics for a nil receiver.
func (e *Error) WithFields(fields map[string]any) *Error {
	derived := *e
	derived.fields = make(map[string]any, len(e.fields)+len(fields))
	maps.Copy(derived.fields, e.fields)
	maps.Copy(derived.fields, fields)

	return &derived
}

// WithCause returns the error with the cause attached. A second cause is joined with those
// attached into one flat MultiError; a nil cause or an empty collection attaches nothing. It
// panics for a nil receiver.
func (e *Error) WithCause(cause error) *Error {
	causes := slices.Concat(members(e.cause), members(cause))

	derived := *e

	switch len(causes) {
	case 0:
		// no-op
	case 1:
		derived.cause = causes[0]
	default:
		derived.cause = Join(causes...)
	}

	return &derived
}

// StackTrace returns the program counters of the stack trace, innermost call first.
func (e *Error) StackTrace() []uintptr {
	if e == nil {
		return nil
	}

	return slices.Clone(e.stack)
}

// Frames returns the call frames of the stack trace, innermost call first.
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

// Format prints the message for %s, %v and %q, the underlying error, the fields sorted by key,
// the stack trace and every cause for %+v, and Go syntax for %#v.
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

// GoString returns the error in Go syntax, without the stack trace.
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

// LogValue returns the error as a slog group of msg, a group of the fields sorted by key, and
// cause if there is one, without the stack trace.
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

	errs, ok := err.(*MultiError)
	if !ok {
		return []error{err}
	}

	var flat []error
	for _, member := range errs.Unwrap() {
		flat = append(flat, members(member)...)
	}

	return flat
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
