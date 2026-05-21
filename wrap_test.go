package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gravitton/assert"
)

func TestWrapUnwrap(t *testing.T) {
	inner := errors.New("inner")
	outer := fmt.Errorf("outer: %w", inner)

	assert.Equal(t, Unwrap(outer), inner)
	assert.Equal(t, Unwrap(inner), nil)
	assert.Equal(t, Unwrap(nil), nil)

	var err *DataError
	assert.Equal(t, Unwrap(err), nil)

	var errs *MultiError
	assert.Equal(t, Unwrap(errs), nil)
}

func TestWrapIs(t *testing.T) {
	inner := errors.New("inner")
	outer := New("outer").WithCause(inner)

	assert.True(t, Is(outer, inner))
	assert.False(t, Is(outer, errors.New("other")))
	assert.False(t, Is(nil, inner))

	var err *DataError
	assert.False(t, Is(err, inner))

	var errs *MultiError
	assert.False(t, Is(errs, inner))
}

func TestWrapAs(t *testing.T) {
	original := New("test").WithField("k", "v")
	wrapped := fmt.Errorf("wrapped: %w", original)

	var target *DataError
	assert.True(t, As(wrapped, &target))
	assert.Equal(t, target, original)

	assert.False(t, As(nil, &target))

	var nilErr *DataError
	assert.True(t, As(nilErr, &target))
	assert.Equal(t, target, nilErr)

	var multiTarget *MultiError
	var nilMulti *MultiError
	assert.True(t, As(nilMulti, &multiTarget))
	assert.Equal(t, multiTarget, nilMulti)
}

func TestWrapAsType(t *testing.T) {
	original := New("test").WithField("k", "v")
	wrapped := fmt.Errorf("wrapped: %w", original)

	target, ok := AsType[*DataError](wrapped)
	assert.True(t, ok)
	assert.Equal(t, target, original)

	_, ok = AsType[*DataError](nil)
	assert.False(t, ok)

	var nilErr *DataError
	nilTarget, ok := AsType[*DataError](nilErr)
	assert.True(t, ok)
	assert.Equal(t, nilTarget, nilErr)
}
