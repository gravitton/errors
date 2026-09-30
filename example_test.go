package errors_test

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"

	"github.com/gravitton/errors"
)

var ErrNotFound = errors.Sentinel("not found")

func newLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if len(groups) == 0 && attr.Key == slog.TimeKey {
				return slog.Attr{}
			}

			return attr
		},
	}))
}

func ExampleNew() {
	err := errors.New("boom")

	fmt.Printf("%v\n", err)
	fmt.Printf("%q\n", err)
	// Output:
	// boom
	// "boom"
}

func ExampleNewf() {
	err := errors.Newf("load user %d: %w", 42, ErrNotFound)

	fmt.Println(err)
	// Output:
	// load user 42: not found
}

func ExampleWrap() {
	inner := errors.New("connection refused")
	err := errors.Wrap(fmt.Errorf("load config: %w", inner))

	fmt.Println(err)
	fmt.Println(slices.Equal(err.StackTrace(), inner.StackTrace()))
	// Output:
	// load config: connection refused
	// true
}

func ExampleError_Fields() {
	err := errors.Wrap(ErrNotFound).WithFields(map[string]any{"id": 42, "table": "users"})
	err = err.WithField("table", "accounts")

	fmt.Println(err)
	fmt.Println(err.Fields())
	// Output:
	// not found
	// map[id:42 table:accounts]
}

func ExampleError_WithCause() {
	cause := errors.New("connection refused").WithField("port", 5432)
	err := errors.New("could not load config").WithCause(cause).WithCause(io.ErrClosedPipe)

	fmt.Println(err)
	fmt.Println(errors.Is(err, io.ErrClosedPipe))
	// Output:
	// could not load config
	// true
}

func ExampleError_Format() {
	err := errors.New("could not load config").WithField("path", "config.yml").WithCause(io.EOF)

	fmt.Printf("%+v\n", err)
	// could not load config
	// 	path=config.yml
	// 	github.com/gravitton/errors_test.ExampleError_Format
	// 		/app/example_test.go:79
	// 	...
	// 	caused by: EOF
}

func ExampleError_GoString() {
	var zero errors.Error
	err := errors.New("boom")

	fmt.Printf("%#v\n", &zero)
	fmt.Printf("%#v\n", err)
	fmt.Printf("%#v\n", err.WithField("id", 42))
	fmt.Printf("%#v\n", err.WithCause(io.EOF))
	// Output:
	// &errors.Error{}
	// &errors.Error{err:&errors.errorString{s:"boom"}}
	// &errors.Error{err:&errors.errorString{s:"boom"}, fields:map[string]interface {}{"id":42}}
	// &errors.Error{err:&errors.errorString{s:"boom"}, cause:&errors.errorString{s:"EOF"}}
}

func ExampleError_LogValue() {
	cause := errors.New("connection refused").WithField("port", 5432)
	err := errors.Wrap(ErrNotFound).WithField("id", 42).WithCause(cause)

	newLogger().Error("load user", "err", err)
	// Output:
	// level=ERROR msg="load user" err.msg="not found" err.fields.id=42 err.cause.msg="connection refused" err.cause.fields.port=5432
}

func ExampleMultiError() {
	errs := errors.NewMulti()
	fmt.Println(errs.ErrorOrNil())

	errs.Add(nil, io.EOF)
	errs.Add(errors.Join(io.ErrUnexpectedEOF, io.ErrClosedPipe))

	fmt.Printf("%v\n", errs)
	// Output:
	// <nil>
	// EOF
	// unexpected EOF
	// io: read/write on closed pipe
}

func ExampleJoin() {
	fmt.Println(errors.Join(nil, nil))
	fmt.Println(errors.Join(io.EOF, nil))
	fmt.Println(errors.Join(io.EOF, errors.New("line one\nline two"), io.ErrClosedPipe))
	// Output:
	// <nil>
	// EOF
	// EOF
	// line one
	// line two
	// io: read/write on closed pipe
}

func ExampleMultiError_Format() {
	errs := errors.Join(errors.New("connection refused").WithField("port", 5432), io.EOF)

	fmt.Printf("%+v\n", errs)
	// connection refused
	// 	port=5432
	// 	github.com/gravitton/errors_test.ExampleMultiError_Format
	// 		/app/example_test.go:143
	// 	...
	// EOF
}

func ExampleMultiError_GoString() {
	fmt.Printf("%#v\n", errors.NewMulti())
	fmt.Printf("%#v\n", errors.Join(io.EOF, errors.New("boom").WithField("id", 42)))
	// Output:
	// &errors.MultiError{}
	// &errors.MultiError{errs:[]error{&errors.errorString{s:"EOF"}, &errors.Error{err:&errors.errorString{s:"boom"}, fields:map[string]interface {}{"id":42}}}}
}

func ExampleMultiError_LogValue() {
	errs := errors.Join(errors.New("connection refused").WithField("port", 5432), io.EOF)

	newLogger().Error("load users", "err", errs)
	// Output:
	// level=ERROR msg="load users" err.0.msg="connection refused" err.0.fields.port=5432 err.1=EOF
}
