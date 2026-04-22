package errors

import (
	"errors"
	"reflect"
	"testing"

	"github.com/gravitton/assert"
)

func TestNew(t *testing.T) {
	err := New("test")

	assert.Equal(t, err.Error(), "test")
	assert.Empty(t, err.Fields())

	cause := err.Unwrap()
	assert.Equal(t, reflect.TypeOf(cause).String(), "*errors.errorString")
}

func TestNewf(t *testing.T) {
	err := Newf("Test DataError #%d: %s", 5, "failed to spawn")

	assert.Equal(t, err.Error(), "Test DataError #5: failed to spawn")
	assert.Empty(t, err.Fields())

	cause := err.Unwrap()
	assert.Equal(t, reflect.TypeOf(cause).String(), "*errors.errorString")
}

func TestWrapNil(t *testing.T) {
	err := Wrap(nil)

	assert.NoError(t, err)
}

func TestWrapNonError(t *testing.T) {
	err := Wrap("something went wrong")

	assert.Equal(t, err.Error(), "something went wrong")

	cause := err.Unwrap()
	assert.Equal(t, reflect.TypeOf(cause).String(), "*errors.errorString")
}

func TestWrapNonErrorNonString(t *testing.T) {
	err := Wrap(159.5)

	assert.Equal(t, err.Error(), "159.5")

	cause := err.Unwrap()
	assert.Equal(t, reflect.TypeOf(cause).String(), "*errors.errorString")
}

func TestWrapError(t *testing.T) {
	original := errors.New("original")
	err := Wrap(original)

	assert.Equal(t, err.Error(), "original")

	cause := err.Unwrap()
	assert.Equal(t, cause, original)
}

func TestWrapDataError(t *testing.T) {
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

func TestWithFieldsDropsFunctions(t *testing.T) {
	err := New("test").WithFields(map[string]any{"key": "value", "func": func() {}})

	assert.Equal(t, err.Fields(), map[string]any{"key": "value"})
}

func TestWithCause(t *testing.T) {
	err1 := New("test")
	original := errors.New("original error")

	err2 := err1.WithCause(original)

	assert.NotSame(t, err1, err2)
	assert.Same(t, err2.Unwrap(), original)
}

func TestErrorsIs(t *testing.T) {
	original := errors.New("original error")
	err1 := New("original error")

	assert.NotErrorIs(t, err1, original)

	err2 := Wrap(original)

	assert.ErrorIs(t, err2, original)

	err3 := err1.WithCause(original)

	assert.ErrorIs(t, err3, original)

	err4 := err2.WithField("action", "call")

	assert.ErrorIs(t, err4, original)
}

func TestErrorsIsDataError(t *testing.T) {
	err1 := New("test")
	err2 := New("test2")
	err3 := New("test").WithFields(map[string]any{"action": "call", "type": "error"})
	err4 := New("test").WithFields(map[string]any{"type": "warn"})

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
