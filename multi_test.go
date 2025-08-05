package errors

import (
	"github.com/gravitton/assert"
	"sync"
	"testing"
)

func TestMultiErrors(t *testing.T) {
	errs := NewMulti()
	assert.NoError(t, errs.ErrorOrNil())
	assert.Length(t, errs.Unwrap(), 0)

	err1 := New("dummy1")
	errs.Append(err1)
	assert.Error(t, errs.ErrorOrNil())
	assert.Length(t, errs.Unwrap(), 1)
	assert.Equal(t, errs.Unwrap(), []error{err1})
	assert.Equal(t, errs.Error(), "dummy1")
	assert.Same(t, errs.Unwrap()[0], err1)
	assert.ErrorIs(t, errs, err1)

	err2 := New("dummy2")
	assert.ErrorIs(t, errs, err1)
	assert.NotErrorIs(t, errs, err2)
	errs.Append(err2)
	assert.Length(t, errs.Unwrap(), 2)
	assert.Equal(t, errs.Unwrap(), []error{err1, err2})
	assert.Equal(t, errs.Error(), "2 errors occurred:\n dummy1\n dummy2")
	assert.ErrorIs(t, errs, err1)
	assert.ErrorIs(t, errs, err2)

	var err3 error
	errs.Append(err3)
	assert.Length(t, errs.Unwrap(), 2)
	assert.Equal(t, errs.Error(), "2 errors occurred:\n dummy1\n dummy2")
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
				errs.Append(Newf("err-%d-%d", i, j))
			}
		}()
	}

	wg.Wait()

	assert.Length(t, errs.Unwrap(), iM*jM)
}
