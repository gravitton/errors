package errors

import (
	"errors"
	"sync"
	"testing"

	"github.com/gravitton/assert"
)

func TestMultiErrorEmpty(t *testing.T) {
	errs := NewMulti()

	assert.Equal(t, errs.Error(), "")
	assert.Equal(t, errs.GoString(), "[]error(nil)")
	assert.Length(t, errs.Unwrap(), 0)
	assert.NoError(t, errs.ErrorOrNil())
}

func TestMultiErrorNil(t *testing.T) {
	var errs *MultiError

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

	errs.Add(err1)
	errs.Add(err2)

	assert.Equal(t, errs.Error(), "2 errors occurred:\n foo\n bar")
	assert.Matches(t, errs.GoString(), `^\[\]error\{(\(\*errors.errorString\)\((0x)?[0-9a-f]+\)(, )?){2}\}$`)
	assert.Length(t, errs.Unwrap(), 2)
	assert.Equal(t, errs.Unwrap(), []error{err1, err2})
	assert.Error(t, errs.ErrorOrNil())
}

func TestMultiErrorAddNil(t *testing.T) {
	errs := NewMulti()

	errs.Add(nil)
	errs.Add(nil)

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

func TestMultiErrorsConcurrentSafe(t *testing.T) {
	errs := NewMulti()

	wg := sync.WaitGroup{}

	iM := 10
	jM := 100

	for i := 0; i < iM; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < jM; j++ {
				errs.Add(Newf("err-%d-%d", i, j))
			}
		}()
	}

	wg.Wait()

	assert.Length(t, errs.Unwrap(), iM*jM)
}
