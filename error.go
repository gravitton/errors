package errors

import (
	"errors"
	"fmt"
	"reflect"
)

type DataError struct {
	err   error
	data  map[string]any
	cause error
}

func New(text string) *DataError {
	return &DataError{
		err: errors.New(text),
	}
}

func Newf(format string, v ...any) *DataError {
	return &DataError{
		err: fmt.Errorf(format, v...),
	}
}

func Wrap(err any) *DataError {
	if err == nil {
		return nil
	}

	var e error
	switch t := err.(type) {
	case *DataError:
		return t
	case error:
		e = t
	default:
		e = fmt.Errorf("%v", err)
	}

	return &DataError{
		err: e,
	}
}

func (e *DataError) Error() string {
	return e.err.Error()
}

func (e *DataError) GoString() string {
	return fmt.Sprintf("%#v %#v", e.err, e.cause)
}

func (e *DataError) Fields() map[string]any {
	return e.data
}

func (e *DataError) WithField(key string, value any) *DataError {
	return e.WithFields(map[string]any{key: value})
}

func (e *DataError) WithFields(values map[string]any) *DataError {
	data := make(map[string]any, len(e.data)+len(values))
	for k, v := range e.data {
		data[k] = v
	}

	for k, v := range values {
		if t := reflect.TypeOf(v); t != nil {
			switch {
			case t.Kind() == reflect.Func, t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Func:
				continue
			}
		}

		data[k] = v
	}

	return &DataError{
		err:   e.err,
		data:  data,
		cause: e.cause,
	}
}

func (e *DataError) WithCause(err error) *DataError {
	return &DataError{
		err:   e.err,
		data:  e.data,
		cause: err,
	}
}

func (e *DataError) Unwrap() error {
	if e.cause != nil {
		return e.cause
	}

	return e.err
}

func (e *DataError) Is(target error) bool {
	err, ok := target.(*DataError)
	if !ok {
		return false
	}

	if e.Error() != target.Error() {
		return false
	}

	for k, v := range err.Fields() {
		if e.Fields()[k] != v {
			return false
		}
	}

	return true
}
