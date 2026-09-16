package errors

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"testing"

	"github.com/gravitton/assert"
)

func TestNew(t *testing.T) {
	err := New("test")

	assert.Equal(t, err.Error(), "test")
	assert.Empty(t, err.Fields())
	assert.Length(t, err.Unwrap(), 1)
}

func TestNewf(t *testing.T) {
	err := Newf("Test Error #%d: %s", 5, "failed to spawn")

	assert.Equal(t, err.Error(), "Test Error #5: failed to spawn")
	assert.Empty(t, err.Fields())
	assert.Length(t, err.Unwrap(), 1)
}

func testMethod(err error) error {
	return Wrap(err)
}

func TestWrapNil(t *testing.T) {
	err1 := Wrap(nil)

	assert.NoError(t, err1)
	assert.True(t, err1 == nil)

	err2 := testMethod(nil)

	assert.NoError(t, err2)
	assert.False(t, err2 == nil) // typed *Error(nil) nil pointer
}

func TestWrapStdError(t *testing.T) {
	original := errors.New("original")
	err := Wrap(original)

	assert.Equal(t, err.Error(), "original")
	assert.Equal(t, err.Unwrap(), []error{original})
}

func TestWrapError(t *testing.T) {
	original := New("original")
	err := Wrap(original)

	assert.Same(t, err, original)
}

func TestStackTrace(t *testing.T) {
	err1 := New("test")
	err2 := Newf("test %d", 1)
	err3 := Wrap(errors.New("std"))

	assert.NotEmpty(t, err1.StackTrace())
	assert.NotEmpty(t, err2.StackTrace())
	assert.NotEmpty(t, err3.StackTrace())
}

func TestFields(t *testing.T) {
	err1 := New("test")

	assert.Empty(t, err1.Fields())

	err2 := err1.WithField("action", "call")

	assert.NotSame(t, err1, err2)
	assert.Empty(t, err1.Fields())
	assert.Equal(t, err2.Fields(), map[string]any{"action": "call"})

	err3 := err2.WithFields(map[string]any{"type": "warning"})

	assert.NotSame(t, err2, err3)
	assert.Equal(t, err2.Fields(), map[string]any{"action": "call"})
	assert.Equal(t, err3.Fields(), map[string]any{"action": "call", "type": "warning"})

	err4 := err3.WithFields(map[string]any{"type": "error", "debug": true, "line": 15})

	assert.NotSame(t, err3, err4)
	assert.Equal(t, err3.Fields(), map[string]any{"action": "call", "type": "warning"})
	assert.Equal(t, err4.Fields(), map[string]any{"action": "call", "type": "error", "debug": true, "line": 15})
}

func TestFieldsAreCopied(t *testing.T) {
	err := New("test").WithField("action", "call")

	fields := err.Fields()
	fields["action"] = "changed"

	assert.Equal(t, err.Fields(), map[string]any{"action": "call"})
}

func TestFieldsUncomparableValues(t *testing.T) {
	base := New("test")

	err1 := base.WithField("tags", []string{"a", "b"})
	err2 := base.WithField("tags", []string{"a", "b"})
	err3 := base.WithField("tags", []string{"a", "c"})
	err4 := base.WithField("tags", "a")

	assert.ErrorIs(t, err1, err2)
	assert.NotErrorIs(t, err1, err3)
	assert.NotErrorIs(t, err1, err4)
}

type container struct {
	value any
}

func TestFieldsUncomparableDynamicValues(t *testing.T) {
	base := New("test")

	err1 := base.WithField("box", container{value: []int{1}})
	err2 := base.WithField("box", container{value: []int{1}})
	err3 := base.WithField("box", container{value: []int{2}})
	err4 := base.WithField("box", container{value: 1})

	assert.ErrorIs(t, err1, err2)
	assert.NotErrorIs(t, err1, err3)
	assert.NotErrorIs(t, err1, err4)
	assert.ErrorIs(t, err4, base.WithField("box", container{value: 1}))
}

func TestFieldsNilValues(t *testing.T) {
	base := New("test")

	assert.ErrorIs(t, base.WithField("k", nil), base.WithField("k", nil))
	assert.NotErrorIs(t, base.WithField("k", nil), base.WithField("k", 1))
	assert.NotErrorIs(t, base.WithField("k", 1), base.WithField("k", nil))
}

func TestFieldsFunctionValues(t *testing.T) {
	base := New("test")
	callback := func() {}

	err1 := base.WithFields(map[string]any{"key": "value", "callback": callback})
	err2 := base.WithField("callback", callback)

	assert.Length(t, err1.Fields(), 2)
	assert.False(t, err1.Is(err2)) // function values never compare equal
}

func TestNilReceiver(t *testing.T) {
	var err *Error

	assert.Equal(t, err.Error(), "<nil>")
	assert.Empty(t, err.Fields())
	assert.Empty(t, err.StackTrace())
	assert.Empty(t, slices.Collect(err.Frames()))
	assert.NoError(t, err.WithField("action", "call"))
	assert.NoError(t, err.WithFields(map[string]any{"action": "call"}))
	assert.NoError(t, err.WithCause(New("cause")))
	assert.Empty(t, err.Unwrap())
	assert.False(t, err.Is(New("test")))
	assert.NotErrorIs(t, New("test"), err)
}

func TestFrames(t *testing.T) {
	frames := slices.Collect(New("test").Frames())

	assert.NotEmpty(t, frames)
	assert.Equal(t, frames[0].Function, "github.com/gravitton/errors.TestFrames")
}

func TestFramesBreak(t *testing.T) {
	count := 0
	for range New("test").Frames() {
		count++

		break
	}

	assert.Equal(t, count, 1)
}

func TestFormat(t *testing.T) {
	err := New("test").WithField("action", "call").WithCause(errors.New("original"))

	assert.Equal(t, fmt.Sprintf("%s", err), "test")
	assert.Equal(t, fmt.Sprintf("%v", err), "test")
	assert.Equal(t, fmt.Sprintf("%q", err), `"test"`)

	details := fmt.Sprintf("%+v", err)

	assert.Contains(t, details, "test")
	assert.Contains(t, details, "action=call")
	assert.Contains(t, details, "caused by: original")
	assert.Contains(t, details, "github.com/gravitton/errors.TestFormat")
	assert.Equal(t, fmt.Sprintf("%d", err), "%!d(string=test)")
}

func TestFormatFlags(t *testing.T) {
	err := New("hello")

	assert.Equal(t, fmt.Sprintf("[%10s]", err), "[     hello]")
	assert.Equal(t, fmt.Sprintf("[%-10v]", err), "[hello     ]")
	assert.Equal(t, fmt.Sprintf("[%.2s]", err), "[he]")
	assert.Equal(t, fmt.Sprintf("%x", err), "68656c6c6f")
	assert.Equal(t, fmt.Sprintf("%#q", err), "`hello`")
}

func TestFormatGoSyntax(t *testing.T) {
	err := New("test").WithField("action", "call").WithCause(io.EOF)

	assert.Equal(t, fmt.Sprintf("%#v", err), `&errors.Error{err:&errors.errorString{s:"test"}, data:map[string]interface {}{"action":"call"}, cause:&errors.errorString{s:"EOF"}}`)

	var nilErr *Error

	assert.Equal(t, fmt.Sprintf("%#v", nilErr), "(*errors.Error)(nil)")
}

func TestFormatNil(t *testing.T) {
	var err *Error

	assert.Equal(t, fmt.Sprintf("%v", err), "<nil>")
	assert.Equal(t, fmt.Sprintf("%+v", err), "<nil>")
}

func TestWithCause(t *testing.T) {
	err1 := New("test")
	original := errors.New("original error")

	err2 := err1.WithCause(original)

	assert.NotSame(t, err1, err2)
	assert.Equal(t, err2.Unwrap(), []error{err1.err, original})
	assert.ErrorIs(t, err2, original)
}

func TestWithCauseKeepsWrappedError(t *testing.T) {
	cause := errors.New("cause")

	err1 := Wrap(io.EOF).WithCause(cause)

	assert.ErrorIs(t, err1, io.EOF)
	assert.ErrorIs(t, err1, cause)

	err2 := Newf("read: %w", io.EOF).WithCause(cause)

	assert.ErrorIs(t, err2, io.EOF)
	assert.ErrorIs(t, err2, cause)
}

func TestErrorsIs(t *testing.T) {
	original := errors.New("original error")
	err1 := New("original error")

	assert.NotErrorIs(t, err1, original)
	assert.NotErrorIs(t, err1, New("original error"))
	assert.ErrorIs(t, Wrap(original), Wrap(original))

	err2 := Wrap(original)

	assert.ErrorIs(t, err2, original)

	err3 := err1.WithCause(original)

	assert.ErrorIs(t, err3, original)

	err4 := err2.WithField("action", "call")

	assert.ErrorIs(t, err4, original)
}

func TestErrorsIsError(t *testing.T) {
	err1 := New("test")
	err2 := New("test2")
	err3 := err1.WithFields(map[string]any{"action": "call", "type": "error"})
	err4 := err1.WithFields(map[string]any{"type": "warn"})

	assert.NotErrorIs(t, err1, err2) // different error
	assert.NotErrorIs(t, err1, err3) // additional fields
	assert.NotErrorIs(t, err1, err4) // additional fields

	assert.NotErrorIs(t, err2, err1) // different error
	assert.NotErrorIs(t, err2, err3) // different error
	assert.NotErrorIs(t, err2, err4) // different error

	assert.ErrorIs(t, err3, err1)    // missing fields
	assert.NotErrorIs(t, err3, err2) // different error
	assert.NotErrorIs(t, err3, err4) // different fields

	assert.ErrorIs(t, err4, err1)    // missing fields
	assert.NotErrorIs(t, err4, err2) // different error
	assert.NotErrorIs(t, err4, err3) // different fields
}

func TestErrorsIsInspectsOnlyTarget(t *testing.T) {
	sentinel := New("sentinel")

	assert.NotErrorIs(t, sentinel, fmt.Errorf("wrapped: %w", sentinel))
	assert.NotErrorIs(t, sentinel, Join(errors.New("other"), sentinel))
	assert.NotErrorIs(t, sentinel, New("other").WithCause(sentinel))
	assert.ErrorIs(t, New("other").WithCause(sentinel), sentinel)
}
