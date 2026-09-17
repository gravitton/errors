package errors

import (
	"errors"
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
	assert.Equal(t, err.Error(), "2 errors occurred:\n foo\n bar")
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

	assert.Equal(t, errs.Error(), "")
	assert.Equal(t, errs.GoString(), "[]error(nil)")
	assert.Equal(t, errs.Len(), 0)
	assert.Length(t, errs.Unwrap(), 0)
	assert.NoError(t, errs.ErrorOrNil())
}

func TestMultiErrorNil(t *testing.T) {
	var errs *MultiError

	assert.Equal(t, errs.Error(), "")
	assert.Equal(t, errs.GoString(), "[]error(nil)")
	assert.Equal(t, errs.Len(), 0)
	assert.Length(t, errs.Unwrap(), 0)
	assert.NoError(t, errs.ErrorOrNil())
}

func TestMultiErrorAddError(t *testing.T) {
	errs := NewMulti()

	err1 := errors.New("foo")
	errs.Add(err1)

	assert.Equal(t, errs.Error(), "foo")
	assert.Matches(t, errs.GoString(), `^\[\]error\{(\(\*errors.errorString\)\((0x)?[0-9a-f]+\)(, )?){1}\}$`)
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
	assert.Equal(t, errs.Error(), "2 errors occurred:\n foo\n bar")
	assert.Matches(t, errs.GoString(), `^\[\]error\{(\(\*errors.errorString\)\((0x)?[0-9a-f]+\)(, )?){2}\}$`)
	assert.Length(t, errs.Unwrap(), 2)
	assert.Equal(t, errs.Unwrap(), []error{err1, err2})
	assert.Error(t, errs.ErrorOrNil())
}

func TestMultiErrorAddNil(t *testing.T) {
	errs := NewMulti()

	errs.Add()
	errs.Add(nil)
	errs.Add(nil, nil)

	assert.Length(t, errs.Unwrap(), 0)
	assert.NoError(t, errs.ErrorOrNil())
}

func TestMultiErrorErrorIs(t *testing.T) {
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

func TestMultiErrorErrorsAreCopied(t *testing.T) {
	err1 := errors.New("foo")
	errs := NewMulti()
	errs.Add(err1)

	unwrapped := errs.Unwrap()
	unwrapped[0] = errors.New("bar")

	assert.Equal(t, errs.Unwrap(), []error{err1})
}

func TestJoinAs(t *testing.T) {
	err := Join(errors.New("foo"), errors.New("bar"))

	errs, ok := AsType[*MultiError](err)

	assert.True(t, ok)
	assert.Equal(t, errs.Len(), 2)
}

func TestMultiErrorsConcurrentSafe(t *testing.T) {
	errs := NewMulti()

	wg := sync.WaitGroup{}

	iM := 10
	jM := 100

	for i := range iM {
		wg.Go(func() {
			for j := range jM {
				errs.Add(Newf("err-%d-%d", i, j))
			}
		})
	}

	wg.Wait()

	assert.Equal(t, errs.Len(), iM*jM)
}

func TestMultiErrorsConcurrentRead(t *testing.T) {
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
			}
		})
	}

	wg.Wait()

	assert.Equal(t, errs.Len(), 1000)
}
