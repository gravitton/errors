package errors

import (
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/gravitton/assert"
)

var errSentinel = Sentinel("sentinel")

var errInit = New("init")

var initFunction = here()

type fsError struct{}

func (e *fsError) Error() string {
	return "fs"
}

type detailedError struct{}

func (e *detailedError) Error() string {
	return "detailed"
}

func (e *detailedError) Format(s fmt.State, verb rune) {
	fmt.Fprint(s, "detailed with details")
}

func here() string {
	pc, _, _, _ := runtime.Caller(1)

	return runtime.FuncForPC(pc).Name()
}

func raisedIn(err *Error) string {
	frame, _ := runtime.CallersFrames(err.StackTrace()).Next()

	return frame.Function
}

func logged(err error) string {
	b := &strings.Builder{}
	handler := slog.NewTextHandler(b, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if len(groups) == 0 && attr.Key != "err" {
				return slog.Attr{}
			}

			return attr
		},
	})

	slog.New(handler).Error("", "err", err)

	return strings.TrimSpace(b.String())
}

func stackAt(err *Error, level int) string {
	b := &strings.Builder{}
	indent := strings.Repeat("\t", level)

	frames := runtime.CallersFrames(err.StackTrace())
	for {
		frame, more := frames.Next()
		fmt.Fprintf(b, "\n%s%s\n%s\t%s:%d", indent, frame.Function, indent, frame.File, frame.Line)

		if !more {
			return b.String()
		}
	}
}

func TestError(t *testing.T) {
	t.Run("zero value is an empty error", func(t *testing.T) {
		var err Error

		assert.Equal(t, err.Error(), "<nil>")
		assert.Equal(t, fmt.Sprintf("%+v", &err), "<nil>")
		assert.Equal(t, fmt.Sprintf("%#v", &err), "&errors.Error{}")
		assert.Empty(t, err.Fields())
		assert.Empty(t, err.StackTrace())
		assert.Empty(t, err.Unwrap())
	})
	t.Run("zero value derives", func(t *testing.T) {
		var err Error

		assert.Equal(t, err.WithField("a", 1).Fields(), map[string]any{"a": 1})
		assert.Empty(t, err.WithField("a", 1).StackTrace())
		assert.Equal(t, fmt.Sprintf("%#v", err.WithField("a", 1)), `&errors.Error{fields:map[string]interface {}{"a":1}}`)
		assert.Same(t, Wrap(&err), &err)
	})
	t.Run("nil receiver is readable", func(t *testing.T) {
		var err *Error

		assert.Equal(t, err.Error(), "<nil>")
		assert.Equal(t, fmt.Sprintf("%v|%+v|%#v", err, err, err), "<nil>|<nil>|(*errors.Error)(nil)")
		assert.Empty(t, err.Unwrap())
		assert.Empty(t, err.Fields())
		assert.Empty(t, err.StackTrace())
		assert.Empty(t, slices.Collect(err.Frames()))
		assert.Equal(t, logged(err), "err=<nil>")
		assert.NotErrorIs(t, err, io.EOF)
		assert.NotErrorIs(t, err, New("test"))
	})
	t.Run("nil receiver panics when built on", func(t *testing.T) {
		var err *Error

		assert.Panics(t, func() {
			err.WithField("a", 1)
		})
		assert.Panics(t, func() {
			err.WithFields(map[string]any{"a": 1})
		})
		assert.Panics(t, func() {
			err.WithCause(io.EOF)
		})
		assert.Panics(t, func() {
			err.WithCause(nil)
		})
	})
}

func TestNew(t *testing.T) {
	t.Run("captures the stack", func(t *testing.T) {
		err := New("test")

		assert.Equal(t, err.Error(), "test")
		assert.Equal(t, raisedIn(err), here())
		assert.Empty(t, err.Fields())
		assert.Length(t, err.Unwrap(), 1)
	})
	t.Run("errors with the same text differ", func(t *testing.T) {
		assert.NotErrorIs(t, New("test"), New("test"))
	})
	t.Run("at package level keeps the init stack", func(t *testing.T) {
		assert.Equal(t, raisedIn(errInit), initFunction)
		assert.Equal(t, raisedIn(errInit.WithField("a", 1)), initFunction)
		assert.Equal(t, raisedIn(Wrap(errInit)), initFunction)
	})
}

func TestSentinel(t *testing.T) {
	t.Run("is a plain error", func(t *testing.T) {
		_, ok := errSentinel.(*Error)

		assert.Equal(t, errSentinel.Error(), "sentinel")
		assert.False(t, ok)
	})
	t.Run("derived errors capture the stack", func(t *testing.T) {
		assert.Equal(t, raisedIn(Wrap(errSentinel).WithField("a", 1)), here())
		assert.Equal(t, raisedIn(Newf("outer: %w", errSentinel)), here())
	})
	t.Run("derived errors are the sentinel", func(t *testing.T) {
		assert.ErrorIs(t, Wrap(errSentinel).WithField("id", 1), errSentinel)
		assert.ErrorIs(t, Newf("outer: %w", errSentinel), errSentinel)
		assert.ErrorIs(t, Wrap(fmt.Errorf("outer: %w", errSentinel)), errSentinel)
		assert.ErrorIs(t, New("outer").WithCause(errSentinel), errSentinel)
		assert.NotErrorIs(t, New("sentinel"), errSentinel)
	})
}

func TestNewf(t *testing.T) {
	t.Run("captures the stack", func(t *testing.T) {
		err := Newf("test %d", 5)

		assert.Equal(t, err.Error(), "test 5")
		assert.Equal(t, raisedIn(err), here())
	})
	t.Run("reuses the stack of a wrapped error", func(t *testing.T) {
		inner := func() *Error {
			return New("inner")
		}()
		err := Newf("outer: %w", inner)

		assert.Equal(t, err.Error(), "outer: inner")
		assert.Equal(t, err.StackTrace(), inner.StackTrace())
	})
}

func TestWrap(t *testing.T) {
	t.Run("nil is nil", func(t *testing.T) {
		assert.True(t, Wrap(nil) == nil)
	})
	t.Run("an error is unchanged", func(t *testing.T) {
		err := New("test")

		assert.Same(t, Wrap(err), err)
	})
	t.Run("a plain error gets the current stack", func(t *testing.T) {
		err := Wrap(io.EOF)

		assert.Equal(t, err.Error(), "EOF")
		assert.Equal(t, err.Unwrap(), []error{io.EOF})
		assert.Equal(t, raisedIn(err), here())
	})
	t.Run("reuses the stack of a wrapped error", func(t *testing.T) {
		inner := func() *Error {
			return New("inner").WithField("a", 1)
		}()
		wrapped := fmt.Errorf("outer: %w", fmt.Errorf("middle: %w", inner))
		err := Wrap(wrapped)

		assert.Equal(t, err.Unwrap(), []error{wrapped})
		assert.Equal(t, err.StackTrace(), inner.StackTrace())
		assert.Empty(t, err.Fields())
	})
	t.Run("a collection gets the current stack", func(t *testing.T) {
		member := func() *Error {
			return New("a")
		}()

		assert.Equal(t, raisedIn(Wrap(Join(io.EOF, member))), here())
		assert.Equal(t, raisedIn(Wrap(fmt.Errorf("%w, %w", io.EOF, member))), here())
	})
	t.Run("a nil error in the chain gets the current stack", func(t *testing.T) {
		var err *Error

		assert.Equal(t, raisedIn(Wrap(fmt.Errorf("outer: %w", err))), here())
		assert.Equal(t, raisedIn(Newf("outer: %w", err)), here())
	})
	t.Run("a struct literal in the chain keeps no stack", func(t *testing.T) {
		assert.Empty(t, Wrap(fmt.Errorf("outer: %w", &Error{})).StackTrace())
	})
}

func TestError_Error(t *testing.T) {
	err := Wrap(fmt.Errorf("outer: %w", io.EOF)).WithField("a", 1).WithCause(io.ErrClosedPipe)

	assert.Equal(t, err.Error(), "outer: EOF")
}

func TestError_Unwrap(t *testing.T) {
	t.Run("underlying error", func(t *testing.T) {
		assert.Equal(t, Wrap(io.EOF).Unwrap(), []error{io.EOF})
	})
	t.Run("underlying error then cause", func(t *testing.T) {
		assert.Equal(t, Wrap(io.EOF).WithCause(io.ErrClosedPipe).Unwrap(), []error{io.EOF, io.ErrClosedPipe})
	})
}

func TestError_Is(t *testing.T) {
	t.Run("copies match the error they came from", func(t *testing.T) {
		err := New("test")

		assert.ErrorIs(t, err.WithField("a", 1), err)
		assert.ErrorIs(t, Wrap(err).WithFields(map[string]any{"a": 1}), err)
		assert.ErrorIs(t, err.WithCause(io.EOF), err)
		assert.ErrorIs(t, errInit.WithField("a", 1), errInit)
	})
	t.Run("matches through the underlying error", func(t *testing.T) {
		assert.ErrorIs(t, Newf("outer: %w", errSentinel), Wrap(errSentinel))
		assert.NotErrorIs(t, Wrap(errSentinel), Newf("outer: %w", errSentinel))
	})
	t.Run("nil target matches nothing", func(t *testing.T) {
		var target *Error

		assert.NotErrorIs(t, New("test"), target)
	})
	t.Run("target without an underlying error matches only itself", func(t *testing.T) {
		var target Error

		assert.ErrorIs(t, &target, &target)
		assert.NotErrorIs(t, target.WithField("a", 1), &target)
		assert.NotErrorIs(t, New("test"), &target)
	})
}

func TestError_Fields(t *testing.T) {
	t.Run("returns a copy", func(t *testing.T) {
		err := New("test").WithField("a", 1)
		err.Fields()["a"] = 2

		assert.Equal(t, err.Fields(), map[string]any{"a": 1})
	})
	t.Run("of a wrapped error stay on it", func(t *testing.T) {
		inner := New("inner").WithField("a", 1)
		err := Newf("outer: %w", inner).WithField("b", 2)

		found, ok := AsType[*Error](fmt.Errorf("handler: %w", inner))

		assert.Equal(t, err.Fields(), map[string]any{"b": 2})
		assert.True(t, ok)
		assert.Equal(t, found.Fields(), map[string]any{"a": 1})
	})
}

func TestError_WithField(t *testing.T) {
	original := New("test").WithCause(io.EOF).WithField("a", 1)
	derived := original.WithField("b", 2)

	assert.Equal(t, original.Fields(), map[string]any{"a": 1})
	assert.Equal(t, derived.Fields(), map[string]any{"a": 1, "b": 2})
	assert.Equal(t, derived.Unwrap(), original.Unwrap())
	assert.Equal(t, derived.StackTrace(), original.StackTrace())
}

func TestError_WithFields(t *testing.T) {
	t.Run("later fields win", func(t *testing.T) {
		err := New("test").WithField("a", 1).WithFields(map[string]any{"a": 2, "b": 3})

		assert.Equal(t, err.Fields(), map[string]any{"a": 2, "b": 3})
	})
	t.Run("copies the map", func(t *testing.T) {
		fields := map[string]any{"a": 1}
		err := New("test").WithFields(fields)
		fields["a"] = 2

		assert.Equal(t, err.Fields(), map[string]any{"a": 1})
	})
}

func TestError_WithCause(t *testing.T) {
	t.Run("attaches the cause", func(t *testing.T) {
		original := New("test").WithField("a", 1)
		err := original.WithCause(io.EOF)

		assert.Equal(t, err.Unwrap(), []error{original.Unwrap()[0], io.EOF})
		assert.Equal(t, err.Fields(), original.Fields())
		assert.Equal(t, err.StackTrace(), original.StackTrace())
		assert.Length(t, original.Unwrap(), 1)
	})
	t.Run("nil, a nil error or an empty collection attaches nothing", func(t *testing.T) {
		err := New("test")
		var missing *Error

		assert.Equal(t, err.WithCause(nil).Unwrap(), err.Unwrap())
		assert.Equal(t, err.WithCause(missing).Unwrap(), err.Unwrap())
		assert.Equal(t, err.WithCause(NewMulti()).Unwrap(), err.Unwrap())
		assert.Same(t, err.WithCause(io.EOF).WithCause(NewMulti()).Unwrap()[1], io.EOF)
	})
	t.Run("a second cause is joined", func(t *testing.T) {
		err := New("test").WithCause(io.EOF).WithCause(io.ErrClosedPipe)

		joined, ok := err.Unwrap()[1].(*MultiError)

		assert.True(t, ok)
		assert.Equal(t, joined.Unwrap(), []error{io.EOF, io.ErrClosedPipe})
	})
	t.Run("causes stay flat", func(t *testing.T) {
		err := New("test").WithCause(io.EOF).WithCause(io.ErrClosedPipe).WithCause(Join(io.ErrShortWrite, io.ErrNoProgress))

		assert.Equal(t, err.Unwrap()[1].(*MultiError).Unwrap(), []error{io.EOF, io.ErrClosedPipe, io.ErrShortWrite, io.ErrNoProgress})
	})
	t.Run("a collection adds a copy of its members", func(t *testing.T) {
		errs := Join(io.EOF, io.ErrClosedPipe).(*MultiError)
		err := New("test").WithCause(errs)
		errs.Add(io.ErrShortWrite)

		assert.Equal(t, err.Unwrap()[1].(*MultiError).Unwrap(), []error{io.EOF, io.ErrClosedPipe})
		assert.Same(t, New("test").WithCause(Join(io.EOF)).Unwrap()[1], io.EOF)
	})
	t.Run("a collection as the first cause is flattened by the next", func(t *testing.T) {
		err := New("test").WithCause(Join(io.EOF, io.ErrClosedPipe)).WithCause(io.ErrShortWrite)

		assert.Equal(t, err.Unwrap()[1].(*MultiError).Unwrap(), []error{io.EOF, io.ErrClosedPipe, io.ErrShortWrite})
	})
	t.Run("derived errors do not share causes", func(t *testing.T) {
		base := New("test").WithCause(io.EOF)
		a := base.WithCause(io.ErrClosedPipe)
		b := base.WithCause(io.ErrUnexpectedEOF)

		assert.Equal(t, a.Unwrap()[1].(*MultiError).Unwrap(), []error{io.EOF, io.ErrClosedPipe})
		assert.Equal(t, b.Unwrap()[1].(*MultiError).Unwrap(), []error{io.EOF, io.ErrUnexpectedEOF})
	})
	t.Run("causes found by Is and As", func(t *testing.T) {
		cause := &fsError{}
		err := New("test").WithCause(cause).WithCause(io.EOF)

		found, ok := AsType[*fsError](err)

		assert.True(t, ok)
		assert.Same(t, found, cause)
		assert.ErrorIs(t, err, io.EOF)
	})
}

func TestError_StackTrace(t *testing.T) {
	err := New("test")
	err.StackTrace()[0] = 0

	assert.Equal(t, raisedIn(err), here())
}

func TestError_Frames(t *testing.T) {
	t.Run("innermost call first", func(t *testing.T) {
		var functions []string
		for frame := range New("test").Frames() {
			functions = append(functions, frame.Function)
			break
		}

		assert.Equal(t, functions, []string{here()})
	})
	t.Run("none without a stack", func(t *testing.T) {
		var err Error

		assert.Empty(t, slices.Collect(err.Frames()))
	})
}

func TestError_Format(t *testing.T) {
	t.Run("message verbs", func(t *testing.T) {
		err := New("test").WithField("a", 1)

		assert.Equal(t, fmt.Sprintf("%s", err), "test")
		assert.Equal(t, fmt.Sprintf("%v", err), "test")
		assert.Equal(t, fmt.Sprintf("%q", err), `"test"`)
		assert.Equal(t, fmt.Sprintf("[%6s]", err), "[  test]")
	})
	t.Run("details with fields, stack and cause", func(t *testing.T) {
		err := New("test").WithField("b", 2).WithField("a", 1).WithCause(io.EOF)

		assert.Equal(t, fmt.Sprintf("%+v", err), "test\n\ta=1\n\tb=2"+stackAt(err, 1)+"\n\tcaused by: EOF")
	})
	t.Run("details of nested causes", func(t *testing.T) {
		cause := func() *Error {
			return New("a").WithCause(io.EOF)
		}()
		err := New("root").WithCause(cause)

		assert.Equal(t, fmt.Sprintf("%+v", err), "root"+stackAt(err, 1)+"\n\tcaused by: a"+stackAt(cause, 2)+"\n\t\tcaused by: EOF")
	})
	t.Run("details of every joined cause", func(t *testing.T) {
		joined := New("root").WithCause(io.EOF).WithCause(io.ErrClosedPipe)
		collection := New("root").WithCause(Join(io.EOF, io.ErrClosedPipe))

		assert.Equal(t, fmt.Sprintf("%+v", joined), "root"+stackAt(joined, 1)+"\n\tcaused by: EOF\n\tcaused by: io: read/write on closed pipe")
		assert.Equal(t, fmt.Sprintf("%+v", collection), "root"+stackAt(collection, 1)+"\n\tcaused by: EOF\n\tcaused by: io: read/write on closed pipe")
	})
	t.Run("details of an error wrapped with %w leave out its fields and cause", func(t *testing.T) {
		inner := func() *Error {
			return New("inner").WithField("a", 1).WithCause(io.EOF)
		}()

		assert.Equal(t, fmt.Sprintf("%+v", Newf("outer: %w", inner)), "outer: inner"+stackAt(inner, 1))
	})
	t.Run("details of the underlying error", func(t *testing.T) {
		err := Wrap(&detailedError{})

		assert.Equal(t, fmt.Sprintf("%+v", err), "detailed with details"+stackAt(err, 1))
	})
	t.Run("details of an underlying collection", func(t *testing.T) {
		member := func() *Error {
			return New("a").WithField("x", 1)
		}()
		err := Wrap(Join(member, io.EOF)).WithField("k", 2)

		assert.Equal(t, fmt.Sprintf("%+v", err), "a\n\tx=1"+stackAt(member, 1)+"\nEOF\n\tk=2"+stackAt(err, 1))
	})
	t.Run("multiline details are indented", func(t *testing.T) {
		err := New("line one\nline two").WithField("a", "value one\nvalue two")

		assert.Equal(t, fmt.Sprintf("%+v", err), "line one\nline two\n\ta=value one\n\tvalue two"+stackAt(err, 1))
	})
}

func TestError_GoString(t *testing.T) {
	t.Run("fields and cause", func(t *testing.T) {
		assert.Equal(t, fmt.Sprintf("%#v", New("test")), `&errors.Error{err:&errors.errorString{s:"test"}}`)
		assert.Equal(t, fmt.Sprintf("%#v", New("test").WithField("a", 1)), `&errors.Error{err:&errors.errorString{s:"test"}, fields:map[string]interface {}{"a":1}}`)
		assert.Equal(t, fmt.Sprintf("%#v", Wrap(errSentinel).WithCause(io.EOF)), `&errors.Error{err:&errors.errorString{s:"sentinel"}, cause:&errors.errorString{s:"EOF"}}`)
	})
	t.Run("foreign wrapper as fmt prints it", func(t *testing.T) {
		assert.Contains(t, fmt.Sprintf("%#v", Newf("outer: %w", io.EOF)), `&errors.Error{err:&fmt.wrapError{msg:"outer: EOF", err:(*errors.errorString)(0x`)
	})
}

func TestError_LogValue(t *testing.T) {
	t.Run("message and fields", func(t *testing.T) {
		assert.Equal(t, logged(New("test").WithField("b", 2).WithField("a", 1)), "err.msg=test err.fields.a=1 err.fields.b=2")
	})
	t.Run("fields do not collide with the message and cause", func(t *testing.T) {
		assert.Equal(t, logged(New("test").WithField("msg", "a").WithField("cause", "b").WithCause(io.EOF)), "err.msg=test err.fields.cause=b err.fields.msg=a err.cause=EOF")
	})
	t.Run("plain cause", func(t *testing.T) {
		assert.Equal(t, logged(New("test").WithCause(io.EOF)), "err.msg=test err.cause=EOF")
	})
	t.Run("cause as a nested group", func(t *testing.T) {
		assert.Equal(t, logged(New("test").WithCause(New("inner").WithField("a", 1))), "err.msg=test err.cause.msg=inner err.cause.fields.a=1")
	})
	t.Run("joined causes keyed by index", func(t *testing.T) {
		err := New("test").WithCause(io.EOF).WithCause(New("inner").WithField("a", 1))

		assert.Equal(t, logged(err), "err.msg=test err.cause.0=EOF err.cause.1.msg=inner err.cause.1.fields.a=1")
	})
}
