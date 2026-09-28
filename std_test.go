package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gravitton/assert"
)

func TestUnwrap(t *testing.T) {
	inner := errors.New("inner")

	assert.Equal(t, Unwrap(fmt.Errorf("outer: %w", inner)), inner)
	assert.Equal(t, Unwrap(inner), nil)
	assert.Equal(t, Unwrap(New("outer").WithCause(inner)), nil)
}

func TestIs(t *testing.T) {
	inner := errors.New("inner")
	outer := New("outer").WithCause(inner)

	assert.True(t, Is(outer, inner))
	assert.False(t, Is(outer, errors.New("other")))
}

func TestAs(t *testing.T) {
	original := New("test")

	var target *Error
	assert.True(t, As(fmt.Errorf("wrapped: %w", original), &target))
	assert.Same(t, target, original)
}

func TestAsType(t *testing.T) {
	original := New("test")

	target, ok := AsType[*Error](fmt.Errorf("wrapped: %w", original))
	assert.True(t, ok)
	assert.Same(t, target, original)
}

func TestErrUnsupported(t *testing.T) {
	assert.Same(t, ErrUnsupported, errors.ErrUnsupported)
}
