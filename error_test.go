package errors

import (
	"errors"
	"reflect"
	"testing"

	"github.com/gravitton/assert"
)

func TestFields(t *testing.T) {
	err := Newf("Test DataError #%d: %s", 5, "failed to spawn")
	assert.Equal(t, err.Error(), "Test DataError #5: failed to spawn")
	assert.Length(t, err.Fields(), 0)

	err = Wrap("test")
	assert.Equal(t, err.Error(), "test")

	err = New("test3").WithField("action", "call")
	assert.Equal(t, err.Error(), "test3")
	assert.Equal(t, err.Fields()["action"], "call")
	assert.NotContains(t, err.Fields(), "type")

	err1 := Wrap(err)
	assert.Same(t, err, err1)

	err2 := err.WithFields(map[string]any{"type": "warning"})
	assert.NotSame(t, err1, err2)
	assert.NotContains(t, err.Fields(), "type")
	assert.Equal(t, err.Error(), "test3")
	assert.Equal(t, err2.Fields()["type"], "warning")

	err3 := err.WithField("action", "send")
	assert.Equal(t, err2.Fields()["action"], "call")
	assert.Equal(t, err3.Fields()["action"], "send")

	err4 := New("error")
	err5 := New("error")
	assert.Equal(t, err4, err5)
	assert.NotSame(t, err4, err5)
}

func TestUnwrap(t *testing.T) {
	err := New("original error 1")

	cause := err.Unwrap()

	assert.Equal(t, reflect.TypeOf(cause).String(), "*errors.errorString")
	assert.Equal(t, "original error 1", cause.Error())

	oErr := errors.New("original error 2")
	err = Wrap(oErr)

	assert.Same(t, oErr, err.Unwrap())

	oErr2 := errors.New("original error 3")
	err = err.WithCause(oErr2)
	assert.Same(t, oErr2, err.Unwrap())
}

func TestErrorsIs(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "*errors.errorString",
			err:  errors.New("dummy error"),
		},
		{
			name: "*DataError",
			err:  New("dummy error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.ErrorIs(t, test.err, test.err)
			assert.ErrorIs(t, Wrap(test.err), test.err)
			assert.ErrorIs(t, Wrap(test.err).WithField("action", "call"), test.err)
			assert.ErrorIs(t, Wrap(test.err).WithField("action", "call").WithField("type", "warning"), test.err)
			assert.ErrorIs(t, Wrap(test.err).WithFields(map[string]any{"type": "warning"}), test.err)
		})
	}
}

func TestErrorsIsWithFields(t *testing.T) {
	assert.ErrorIs(t, New("dummy error").WithField("action", "call"), New("dummy error"))
	assert.NotErrorIs(t, New("dummy error"), New("dummy error").WithField("action", "call"))
	assert.NotErrorIs(t, New("dummy error").WithField("action", "call"), New("dummy error").WithField("action", "send"))

	err := New("dummy error").WithField("module", "http")
	assert.ErrorIs(t, err.WithField("add", false), err)
	assert.NotErrorIs(t, err, err.WithField("add", false))
}
