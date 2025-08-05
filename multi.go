package errors

import (
	"fmt"
	"strings"
	"sync"
)

type MultiError struct {
	errs  []error
	mutex sync.Mutex
}

func NewMulti() *MultiError {
	return &MultiError{}
}

func (e *MultiError) Error() string {
	if len(e.errs) == 1 {
		return e.errs[0].Error()
	}

	msg := make([]string, len(e.errs))
	for i, err := range e.errs {
		msg[i] = err.Error()
	}

	return fmt.Sprintf("%d errors occurred:\n %s", len(e.errs), strings.Join(msg, "\n "))
}

func (e *MultiError) GoString() string {
	return fmt.Sprintf("%#v", e.errs)
}

func (e *MultiError) Unwrap() []error {
	if e == nil {
		return nil
	}

	return e.errs
}

func (e *MultiError) Append(err error) {
	if err == nil {
		return
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.errs = append(e.errs, err)
}

func (e *MultiError) ErrorOrNil() error {
	if len(e.errs) == 0 {
		return nil
	}

	return e
}
