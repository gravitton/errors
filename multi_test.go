package errors

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"

	"github.com/gravitton/assert"
)

func TestJoinNone(t *testing.T) {
	err := Join()

	assert.NoError(t, err)
}

func TestJoinAllNil(t *testing.T) {
	err := Join(nil, nil)

	assert.NoError(t, err)
}

func TestJoinSingle(t *testing.T) {
	err1 := errors.New("foo")
	err := Join(err1)

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "foo")
	assert.ErrorIs(t, err, err1)
}

func TestJoinMultiple(t *testing.T) {
	err1 := errors.New("foo")
	err2 := errors.New("bar")
	err := Join(err1, err2)

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "2 errors occurred:\n\t1. foo\n\t2. bar")
	assert.ErrorIs(t, err, err1)
	assert.ErrorIs(t, err, err2)
}

func TestJoinSkipsNils(t *testing.T) {
	err1 := errors.New("foo")
	err := Join(nil, err1, nil)

	assert.Error(t, err)
	assert.Equal(t, err.Error(), "foo")
}

func TestMultiErrorEmpty(t *testing.T) {
	errs := NewMulti()

	assert.Equal(t, errs.Error(), "no errors")
	assert.Equal(t, errs.GoString(), "&errors.MultiError{errs:[]error(nil)}")
	assert.Equal(t, errs.Len(), 0)
	assert.Length(t, errs.Unwrap(), 0)
	assert.NoError(t, errs.ErrorOrNil())
}

func TestMultiErrorAddError(t *testing.T) {
	errs := NewMulti()

	err1 := errors.New("foo")
	errs.Add(err1)

	assert.Equal(t, errs.Error(), "foo")
	assert.Equal(t, errs.GoString(), `&errors.MultiError{errs:[]error{&errors.errorString{s:"foo"}}}`)
	assert.Length(t, errs.Unwrap(), 1)
	assert.Equal(t, errs.Unwrap(), []error{err1})
	assert.Error(t, errs.ErrorOrNil())
}

func TestMultiErrorAddErrors(t *testing.T) {
	errs := NewMulti()

	err1 := errors.New("foo")
	err2 := errors.New("bar")

	errs.Add(err1, err2)

	assert.Equal(t, errs.Len(), 2)
	assert.Equal(t, errs.Error(), "2 errors occurred:\n\t1. foo\n\t2. bar")
	assert.Equal(t, errs.GoString(), `&errors.MultiError{errs:[]error{&errors.errorString{s:"foo"}, &errors.errorString{s:"bar"}}}`)
	assert.Length(t, errs.Unwrap(), 2)
	assert.Equal(t, errs.Unwrap(), []error{err1, err2})
	assert.Error(t, errs.ErrorOrNil())
}

func TestMultiErrorErrorOrNilIsCollection(t *testing.T) {
	errs := NewMulti()
	errs.Add(io.EOF)

	err := errs.ErrorOrNil()
	errs.Add(io.ErrClosedPipe)

	assert.Same(t, err, error(errs))
	assert.ErrorIs(t, err, io.ErrClosedPipe)
}

func TestMultiErrorAddNil(t *testing.T) {
	errs := NewMulti()

	errs.Add()
	errs.Add(nil)
	errs.Add(nil, nil)

	assert.Length(t, errs.Unwrap(), 0)
	assert.NoError(t, errs.ErrorOrNil())
}

func TestMultiErrorMultilineMessage(t *testing.T) {
	errs := Join(errors.New("a\nb"), io.EOF)

	assert.Equal(t, errs.Error(), "2 errors occurred:\n\t1. a\n\t\tb\n\t2. EOF")
}

func TestMultiErrorEmptyLines(t *testing.T) {
	errs := Join(errors.New("a\n\nb"), errors.New("c\n"))

	assert.Equal(t, errs.Error(), "2 errors occurred:\n\t1. a\n\n\t\tb\n\t2. c\n")
}

func TestMultiErrorNested(t *testing.T) {
	errs := Join(io.EOF, Join(io.ErrUnexpectedEOF, io.ErrClosedPipe))

	assert.Equal(t, errs.Error(), "2 errors occurred:\n\t1. EOF\n\t2. 2 errors occurred:\n\t\t1. unexpected EOF\n\t\t2. io: read/write on closed pipe")
}

func TestMultiErrorIs(t *testing.T) {
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
}

func TestMultiErrorUnwrapAppend(t *testing.T) {
	errs := NewMulti()
	errs.Add(io.EOF)
	errs.Add(io.ErrUnexpectedEOF)
	errs.Add(io.ErrShortWrite)

	unwrapped := errs.Unwrap()
	errs.Add(io.ErrClosedPipe)
	_ = append(unwrapped, io.ErrNoProgress)

	assert.Length(t, unwrapped, 3)
	assert.Equal(t, errs.Unwrap(), []error{io.EOF, io.ErrUnexpectedEOF, io.ErrShortWrite, io.ErrClosedPipe})
}

func TestMultiErrorFormat(t *testing.T) {
	err1 := New("foo").WithField("k", 1)
	err2 := io.EOF
	errs := Join(err1, err2)

	assert.Equal(t, fmt.Sprintf("%s", errs), "2 errors occurred:\n\t1. foo\n\t2. EOF")
	assert.Equal(t, fmt.Sprintf("%q", Join(err2)), `"EOF"`)
	assert.Equal(t, fmt.Sprintf("[%5s]", Join(err2)), "[  EOF]")
	assert.Equal(t, fmt.Sprintf("%#v", Join(err2)), `&errors.MultiError{errs:[]error{&errors.errorString{s:"EOF"}}}`)
}

func TestMultiErrorFormatDetails(t *testing.T) {
	err1 := New("foo").WithField("k", 1).WithCause(io.EOF)
	errs := Join(err1, errors.New("a\nb"))

	assert.Equal(t, fmt.Sprintf("%+v", errs), "2 errors occurred:\n\t1. foo\n\t\tk=1"+stackAt(err1, 2)+"\n\t\tcaused by: EOF\n\t2. a\n\t\tb")
}

func TestMultiErrorFormatSingle(t *testing.T) {
	err := New("foo").WithField("k", 1)

	assert.Equal(t, fmt.Sprintf("%+v", Join(err)), fmt.Sprintf("%+v", err))
}

func TestMultiErrorFormatEmpty(t *testing.T) {
	assert.Equal(t, fmt.Sprintf("%+v", NewMulti()), "no errors")
	assert.Equal(t, fmt.Sprintf("%v", NewMulti()), "no errors")
}

func TestJoinAs(t *testing.T) {
	err := Join(errors.New("foo"), errors.New("bar"))

	errs, ok := AsType[*MultiError](err)

	assert.True(t, ok)
	assert.Equal(t, errs.Len(), 2)
}

func TestJoinAsMember(t *testing.T) {
	member := &fsError{}
	err := Join(errors.New("foo"), member)

	found, ok := AsType[*fsError](err)

	assert.True(t, ok)
	assert.Same(t, found, member)
}

func TestMultiErrorConcurrentAdd(t *testing.T) {
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
}

func TestMultiErrorConcurrentRead(t *testing.T) {
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
}
