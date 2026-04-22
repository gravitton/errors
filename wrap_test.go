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
}

func TestWrapIs(t *testing.T) {
	inner := errors.New("inner")
	outer := New("outer").WithCause(inner)

	assert.True(t, Is(outer, inner))
	assert.False(t, Is(outer, errors.New("other")))
}

func TestWrapAs(t *testing.T) {
	original := New("test").WithField("k", "v")
	wrapped := fmt.Errorf("wrapped: %w", original)

	var target *DataError
	assert.True(t, As(wrapped, &target))
	assert.Equal(t, target, original)
}
