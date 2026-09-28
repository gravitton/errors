package errors

import (
	"fmt"
	"io"
	"runtime"
	"strings"
	"testing"

	"github.com/gravitton/assert"
)

var errSentinel = Sentinel("sentinel")

type fsError struct{}

func (e *fsError) Error() string {
	return "fs"
}

type uncomparableError []string

func (e uncomparableError) Error() string {
	return "uncomparable"
}

type detailedError struct{}

func (e *detailedError) Error() string {
	return "detailed"
}

func (e *detailedError) Format(s fmt.State, verb rune) {
	fmt.Fprint(s, "detailed with details")
}

func raisedIn(err *Error) string {
	frame, _ := runtime.CallersFrames(err.StackTrace()).Next()

	return strings.TrimPrefix(frame.Function, "github.com/gravitton/errors.")
}

func TestNew(t *testing.T) {
	err := New("test")

	assert.Equal(t, err.Error(), "test")
	assert.Equal(t, raisedIn(err), "TestNew")
	assert.Empty(t, err.Fields())
	assert.Length(t, err.Unwrap(), 1)
}

func TestNewDistinct(t *testing.T) {
	assert.NotErrorIs(t, New("test"), New("test"))
}

func TestSentinel(t *testing.T) {
	assert.Equal(t, errSentinel.Error(), "sentinel")
	assert.Empty(t, errSentinel.StackTrace())
}

func TestNewf(t *testing.T) {
	err := Newf("test %d", 5)

	assert.Equal(t, err.Error(), "test 5")
	assert.Equal(t, raisedIn(err), "TestNewf")
}

func TestNewfReusesStack(t *testing.T) {
	inner := New("inner")
	err := func() *Error {
		return Newf("outer: %w", inner)
	}()

	assert.Equal(t, err.Error(), "outer: inner")
	assert.Equal(t, err.StackTrace(), inner.StackTrace())
}

func TestWrapNil(t *testing.T) {
	assert.True(t, Wrap(nil) == nil)
}

func TestWrapErrorWithStack(t *testing.T) {
	err := New("test")

	assert.Same(t, Wrap(err), err)
}

func TestWrapSentinel(t *testing.T) {
	err := Wrap(errSentinel)

	assert.Equal(t, raisedIn(err), "TestWrapSentinel")
	assert.ErrorIs(t, err, errSentinel)
	assert.Empty(t, errSentinel.StackTrace())
}

func TestWrapStdError(t *testing.T) {
	err := Wrap(io.EOF)

	assert.Equal(t, err.Error(), "EOF")
	assert.Equal(t, err.Unwrap(), []error{io.EOF})
	assert.Equal(t, raisedIn(err), "TestWrapStdError")
}

func TestWrapReusesStack(t *testing.T) {
	inner := New("inner")
	err := func() *Error {
		return Wrap(fmt.Errorf("outer: %w", inner))
	}()

	assert.Equal(t, err.StackTrace(), inner.StackTrace())
}

func TestWrapSkipsCollections(t *testing.T) {
	err := Wrap(Join(New("a").WithField("a", 1), New("b")))

	assert.Equal(t, raisedIn(err), "TestWrapSkipsCollections")
	assert.Empty(t, err.Fields())
}

func TestWrapSkipsCauses(t *testing.T) {
	err := Wrap(fmt.Errorf("outer: %w", errSentinel.WithCause(New("cause").WithField("a", 1))))

	assert.Equal(t, raisedIn(err), "TestWrapSkipsCauses")
	assert.Empty(t, err.Fields())
}

func TestWithFieldIsImmutable(t *testing.T) {
	original := New("test").WithField("a", 1)
	derived := original.WithField("b", 2)

	assert.Equal(t, original.Fields(), map[string]any{"a": 1})
	assert.Equal(t, derived.Fields(), map[string]any{"a": 1, "b": 2})
	assert.Equal(t, derived.StackTrace(), original.StackTrace())
}

func TestWithFields(t *testing.T) {
	err := New("test").WithField("a", 1).WithFields(map[string]any{"a": 2, "b": 3})

	assert.Equal(t, err.Fields(), map[string]any{"a": 2, "b": 3})
}

func TestWithFieldsCopiesMap(t *testing.T) {
	fields := map[string]any{"a": 1}
	err := New("test").WithFields(fields)
	fields["a"] = 2

	assert.Equal(t, err.Fields(), map[string]any{"a": 1})
}

func TestDerivedSentinelCapturesStack(t *testing.T) {
	assert.Equal(t, raisedIn(errSentinel.WithField("a", 1)), "TestDerivedSentinelCapturesStack")
	assert.Equal(t, raisedIn(errSentinel.WithFields(map[string]any{"a": 1})), "TestDerivedSentinelCapturesStack")
	assert.Equal(t, raisedIn(errSentinel.WithCause(io.EOF)), "TestDerivedSentinelCapturesStack")
	assert.Empty(t, errSentinel.StackTrace())
}

func TestFieldsMergeMainChain(t *testing.T) {
	inner := New("inner").WithFields(map[string]any{"a": 1, "b": 1})
	err := Wrap(fmt.Errorf("outer: %w", inner)).WithField("b", 2)

	assert.Equal(t, err.Fields(), map[string]any{"a": 1, "b": 2})
}

func TestFieldsAreCopied(t *testing.T) {
	err := New("test").WithField("a", 1)
	err.Fields()["a"] = 2

	assert.Equal(t, err.Fields(), map[string]any{"a": 1})
}

func TestWithCause(t *testing.T) {
	original := New("test")
	err := original.WithCause(io.EOF).WithCause(io.ErrClosedPipe)

	assert.Equal(t, err.Unwrap(), []error{original.Unwrap()[0], io.EOF, io.ErrClosedPipe})
	assert.Equal(t, err.StackTrace(), original.StackTrace())
	assert.Length(t, original.Unwrap(), 1)
	assert.ErrorIs(t, err, io.EOF)
	assert.ErrorIs(t, err, io.ErrClosedPipe)
}

func TestWithCauseNil(t *testing.T) {
	err := New("test")

	assert.Same(t, err.WithCause(nil), err)
}

func TestWithCauseDoesNotShareCauses(t *testing.T) {
	base := New("test").WithCause(io.EOF)
	a := base.WithCause(io.ErrClosedPipe)
	b := base.WithCause(io.ErrUnexpectedEOF)

	assert.Equal(t, a.Unwrap()[2], io.ErrClosedPipe)
	assert.Equal(t, b.Unwrap()[2], io.ErrUnexpectedEOF)
}

func TestAsFindsCause(t *testing.T) {
	cause := &fsError{}
	err := New("test").WithCause(cause)

	found, ok := AsType[*fsError](err)

	assert.True(t, ok)
	assert.Same(t, found, cause)
}

func TestIsSentinelDerived(t *testing.T) {
	assert.ErrorIs(t, errSentinel, errSentinel)
	assert.ErrorIs(t, errSentinel.WithField("id", 1), errSentinel)
	assert.ErrorIs(t, errSentinel.WithCause(io.EOF), errSentinel)
	assert.ErrorIs(t, Wrap(errSentinel), errSentinel)
	assert.ErrorIs(t, Newf("outer: %w", errSentinel), errSentinel)
	assert.ErrorIs(t, New("outer").WithCause(errSentinel), errSentinel)
	assert.NotErrorIs(t, New("sentinel"), errSentinel)
}

func TestIsIgnoresFields(t *testing.T) {
	err := errSentinel.WithField("id", 42)

	assert.ErrorIs(t, err, errSentinel.WithField("id", 7))
	assert.ErrorIs(t, errSentinel, err)
}

func TestIsUnderlyingChain(t *testing.T) {
	assert.ErrorIs(t, Newf("outer: %w", io.EOF), Wrap(io.EOF))
	assert.NotErrorIs(t, Wrap(io.EOF), Newf("outer: %w", io.EOF))
}

func TestIsUncomparableUnderlying(t *testing.T) {
	assert.NotErrorIs(t, Wrap(uncomparableError{}), Wrap(uncomparableError{}))
}

func TestIsStdError(t *testing.T) {
	assert.ErrorIs(t, Wrap(io.EOF), io.EOF)
	assert.ErrorIs(t, Wrap(fmt.Errorf("outer: %w", io.EOF)), io.EOF)
}

func TestFormat(t *testing.T) {
	err := New("test").WithField("a", 1)

	assert.Equal(t, fmt.Sprintf("%s", err), "test")
	assert.Equal(t, fmt.Sprintf("%v", err), "test")
	assert.Equal(t, fmt.Sprintf("%q", err), `"test"`)
	assert.Equal(t, fmt.Sprintf("[%6s]", err), "[  test]")
	assert.Equal(t, fmt.Sprintf("%#v", err), `&errors.Error{err:&errors.errorString{s:"test"}, fields:map[string]interface {}{"a":1}, causes:[]error{}}`)
	assert.Equal(t, fmt.Sprintf("%#v", errSentinel.WithCause(io.EOF)), `&errors.Error{err:&errors.errorString{s:"sentinel"}, fields:map[string]interface {}(nil), causes:[]error{&errors.errorString{s:"EOF"}}}`)
}

func TestFormatDetails(t *testing.T) {
	err := New("test").WithField("b", 2).WithField("a", 1).WithCause(io.EOF).WithCause(io.ErrClosedPipe)
	details := fmt.Sprintf("%+v", err)

	assert.True(t, strings.HasPrefix(details, "test\n\ta=1\n\tb=2\n\tgithub.com/gravitton/errors.TestFormatDetails\n\t\t"))
	assert.True(t, strings.HasSuffix(details, "\ncaused by: EOF\ncaused by: io: read/write on closed pipe"))
}

func TestFormatDetailsNestedCauses(t *testing.T) {
	err := New("root").WithCause(New("a").WithCause(io.EOF)).WithCause(io.ErrClosedPipe)
	details := fmt.Sprintf("%+v", err)

	assert.True(t, strings.Contains(details, "\ncaused by: a\n\t\tgithub.com/gravitton/errors.TestFormatDetailsNestedCauses\n\t\t\t"))
	assert.True(t, strings.HasSuffix(details, "\n\tcaused by: EOF\ncaused by: io: read/write on closed pipe"))
}

func TestFormatDetailsNested(t *testing.T) {
	inner := New("inner").WithField("a", 1).WithCause(io.EOF)
	details := fmt.Sprintf("%+v", Newf("outer: %w", inner))

	assert.True(t, strings.HasPrefix(details, "outer: inner\n\ta=1\n\tgithub.com/gravitton/errors.TestFormatDetailsNested\n"))
	assert.Equal(t, strings.Count(details, "a=1"), 1)
	assert.Equal(t, strings.Count(details, "TestFormatDetailsNested"), 1)
	assert.True(t, strings.HasSuffix(details, "\ncaused by: EOF"))
}

func TestFormatDetailsUnderlyingMessageOnly(t *testing.T) {
	details := fmt.Sprintf("%+v", Wrap(&detailedError{}))

	assert.True(t, strings.HasPrefix(details, "detailed\n\tgithub.com/gravitton/errors.TestFormatDetailsUnderlyingMessageOnly\n"))
}

func TestFormatDetailsSentinel(t *testing.T) {
	assert.Equal(t, fmt.Sprintf("%+v", errSentinel), "sentinel")
}
