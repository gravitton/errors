package errors

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"

	"github.com/gravitton/assert"
)

func TestMultiError(t *testing.T) {
	t.Run("zero value is usable", func(t *testing.T) {
		var errs MultiError
		errs.Add(io.EOF)

		assert.Equal(t, errs.Len(), 1)
		assert.Equal(t, errs.Error(), "EOF")
	})
	t.Run("concurrent adds are all kept", func(t *testing.T) {
		errs := NewMulti()

		wg := sync.WaitGroup{}

		for i := range 10 {
			wg.Go(func() {
				for j := range 100 {
					errs.Add(Newf("err-%d-%d", i, j))
				}
			})
		}

		wg.Wait()

		assert.Equal(t, errs.Len(), 1000)
	})
	t.Run("concurrent reads while adding", func(t *testing.T) {
		errs := NewMulti()

		wg := sync.WaitGroup{}

		for i := range 10 {
			wg.Go(func() {
				for j := range 100 {
					errs.Add(Newf("err-%d-%d", i, j))
				}
			})

			wg.Go(func() {
				for range 100 {
					_ = errs.Error()
					_ = errs.GoString()
					_ = errs.ErrorOrNil()
					_ = append(errs.Unwrap(), io.EOF)
				}
			})
		}

		wg.Wait()

		assert.Equal(t, errs.Len(), 1000)
	})
}

func TestNewMulti(t *testing.T) {
	errs := NewMulti()

	assert.Equal(t, errs.Error(), "")
	assert.Equal(t, errs.Len(), 0)
	assert.Length(t, errs.Unwrap(), 0)
	assert.NoError(t, errs.ErrorOrNil())
}

func TestJoin(t *testing.T) {
	t.Run("no errors is nil", func(t *testing.T) {
		assert.NoError(t, Join())
	})
	t.Run("all nil is nil", func(t *testing.T) {
		assert.NoError(t, Join(nil, nil))
	})
	t.Run("single error keeps its message", func(t *testing.T) {
		err1 := errors.New("foo")
		err := Join(err1)

		assert.Error(t, err)
		assert.Equal(t, err.Error(), "foo")
		assert.ErrorIs(t, err, err1)
	})
	t.Run("messages joined by newlines", func(t *testing.T) {
		err1 := errors.New("foo")
		err2 := errors.New("bar")
		err := Join(err1, err2)

		assert.Error(t, err)
		assert.Equal(t, err.Error(), "foo\nbar")
		assert.ErrorIs(t, err, err1)
		assert.ErrorIs(t, err, err2)
	})
	t.Run("skips nils", func(t *testing.T) {
		err1 := errors.New("foo")
		err := Join(nil, err1, nil)

		assert.Error(t, err)
		assert.Equal(t, err.Error(), "foo")
	})
	t.Run("collection found by As", func(t *testing.T) {
		err := Join(errors.New("foo"), errors.New("bar"))

		errs, ok := AsType[*MultiError](err)

		assert.True(t, ok)
		assert.Equal(t, errs.Len(), 2)
	})
	t.Run("member found by As", func(t *testing.T) {
		member := &fsError{}
		err := Join(errors.New("foo"), member)

		found, ok := AsType[*fsError](err)

		assert.True(t, ok)
		assert.Same(t, found, member)
	})
}

func TestMultiError_Add(t *testing.T) {
	t.Run("one error", func(t *testing.T) {
		errs := NewMulti()

		err1 := errors.New("foo")
		errs.Add(err1)

		assert.Equal(t, errs.Unwrap(), []error{err1})
		assert.Error(t, errs.ErrorOrNil())
	})
	t.Run("several errors in order", func(t *testing.T) {
		errs := NewMulti()

		err1 := errors.New("foo")
		err2 := errors.New("bar")

		errs.Add(err1, err2)

		assert.Equal(t, errs.Len(), 2)
		assert.Equal(t, errs.Unwrap(), []error{err1, err2})
		assert.Error(t, errs.ErrorOrNil())
	})
	t.Run("skips nils", func(t *testing.T) {
		errs := NewMulti()

		errs.Add()
		errs.Add(nil)
		errs.Add(nil, nil)

		assert.Length(t, errs.Unwrap(), 0)
		assert.NoError(t, errs.ErrorOrNil())
	})
}

func TestMultiError_Error(t *testing.T) {
	t.Run("messages joined by newlines", func(t *testing.T) {
		assert.Equal(t, Join(errors.New("foo"), errors.New("bar")).Error(), "foo\nbar")
	})
	t.Run("nested collections are flat text", func(t *testing.T) {
		errs := Join(io.EOF, Join(io.ErrUnexpectedEOF, io.ErrClosedPipe))

		assert.Equal(t, errs.Error(), "EOF\nunexpected EOF\nio: read/write on closed pipe")
	})
}

func TestMultiError_Unwrap(t *testing.T) {
	t.Run("members visible to Is", func(t *testing.T) {
		errs := NewMulti()

		err1 := errors.New("foo")
		err2 := errors.New("bar")

		assert.NotErrorIs(t, errs, err1)
		assert.NotErrorIs(t, errs, err2)

		errs.Add(err1)

		assert.ErrorIs(t, errs, err1)
		assert.NotErrorIs(t, errs, err2)

		errs.Add(err2)

		assert.ErrorIs(t, errs, err1)
		assert.ErrorIs(t, errs, err2)
	})
	t.Run("appending does not change the collection", func(t *testing.T) {
		errs := NewMulti()
		errs.Add(io.EOF)
		errs.Add(io.ErrUnexpectedEOF)
		errs.Add(io.ErrShortWrite)

		unwrapped := errs.Unwrap()
		errs.Add(io.ErrClosedPipe)
		_ = append(unwrapped, io.ErrNoProgress)

		assert.Length(t, unwrapped, 3)
		assert.Equal(t, errs.Unwrap(), []error{io.EOF, io.ErrUnexpectedEOF, io.ErrShortWrite, io.ErrClosedPipe})
	})
}

func TestMultiError_ErrorOrNil(t *testing.T) {
	errs := NewMulti()
	errs.Add(io.EOF)

	err := errs.ErrorOrNil()
	errs.Add(io.ErrClosedPipe)

	assert.Same(t, err, error(errs))
	assert.ErrorIs(t, err, io.ErrClosedPipe)
}

func TestMultiError_Format(t *testing.T) {
	t.Run("message verbs", func(t *testing.T) {
		errs := Join(New("foo").WithField("k", 1), io.EOF)

		assert.Equal(t, fmt.Sprintf("%s", errs), "foo\nEOF")
		assert.Equal(t, fmt.Sprintf("%q", Join(io.EOF)), `"EOF"`)
		assert.Equal(t, fmt.Sprintf("[%5s]", Join(io.EOF)), "[  EOF]")
	})
	t.Run("details of members joined by newlines", func(t *testing.T) {
		err1 := New("foo").WithField("k", 1).WithCause(io.EOF)
		errs := Join(err1, errors.New("a\nb"))

		assert.Equal(t, fmt.Sprintf("%+v", errs), "foo\n\tk=1"+stackAt(err1, 1)+"\n\tcaused by: EOF\na\nb")
	})
	t.Run("empty collection prints nothing", func(t *testing.T) {
		assert.Equal(t, fmt.Sprintf("%+v", NewMulti()), "")
		assert.Equal(t, fmt.Sprintf("%v", NewMulti()), "")
	})
}

func TestMultiError_GoString(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		assert.Equal(t, fmt.Sprintf("%#v", NewMulti()), "&errors.MultiError{errs:[]error(nil)}")
	})
	t.Run("members in go syntax", func(t *testing.T) {
		assert.Equal(t, fmt.Sprintf("%#v", Join(errors.New("foo"), io.EOF)), `&errors.MultiError{errs:[]error{&errors.errorString{s:"foo"}, &errors.errorString{s:"EOF"}}}`)
	})
}
