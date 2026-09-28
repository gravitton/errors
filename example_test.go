package errors_test

import (
	"fmt"
	"io"

	"github.com/gravitton/errors"
)

var ErrNotFound = errors.Sentinel("not found")

func ExampleNew() {
	err := errors.New("boom")

	fmt.Printf("%v\n", err)
	fmt.Printf("%q\n", err)
	fmt.Printf("%#v\n", err)
	// Output:
	// boom
	// "boom"
	// &errors.Error{err:&errors.errorString{s:"boom"}, fields:map[string]interface {}(nil), causes:[]error{}}
}

func ExampleNewf() {
	err := errors.Newf("load user %d: %w", 42, ErrNotFound)

	fmt.Println(err)
	// Output:
	// load user 42: not found
}

func ExampleWrap() {
	err := errors.Wrap(io.EOF)

	fmt.Printf("%v\n", err)
	fmt.Printf("%#v\n", err)
	// Output:
	// EOF
	// &errors.Error{err:&errors.errorString{s:"EOF"}, fields:map[string]interface {}(nil), causes:[]error{}}
}

func ExampleError_Fields() {
	err := ErrNotFound.WithFields(map[string]any{"id": 42, "table": "users"})
	err = errors.Newf("load user: %w", err)
	err = errors.Wrap(fmt.Errorf("handler: %w", err)).WithField("table", "accounts")

	fmt.Println(err)
	fmt.Println(err.Fields())
	// Output:
	// handler: load user: not found
	// map[id:42 table:accounts]
}

func ExampleError_WithCause() {
	cause := errors.New("connection refused").WithField("port", 5432)
	err := errors.New("could not load config").WithCause(cause).WithCause(io.ErrClosedPipe)

	fmt.Printf("%v\n", err)
	fmt.Printf("%#v\n", err)
	// Output:
	// could not load config
	// &errors.Error{err:&errors.errorString{s:"could not load config"}, fields:map[string]interface {}(nil), causes:[]error{&errors.Error{err:&errors.errorString{s:"connection refused"}, fields:map[string]interface {}{"port":5432}, causes:[]error{}}, &errors.errorString{s:"io: read/write on closed pipe"}}}
}

func ExampleJoin() {
	fmt.Println(errors.Join(nil, nil))
	fmt.Println(errors.Join(io.EOF, nil))
	fmt.Println(errors.Join(io.EOF, errors.New("line one\nline two"), io.ErrClosedPipe))
	// Output:
	// <nil>
	// EOF
	// 3 errors occurred:
	// 	EOF
	// 	line one
	// 	line two
	// 	io: read/write on closed pipe
}

func ExampleMultiError() {
	errs := errors.NewMulti()
	fmt.Println(errs.ErrorOrNil())

	errs.Add(nil, io.EOF)
	errs.Add(errors.Join(io.ErrUnexpectedEOF, io.ErrClosedPipe))

	fmt.Printf("%v\n", errs)
	fmt.Printf("%#v\n", errs)
	// Output:
	// <nil>
	// 2 errors occurred:
	// 	EOF
	// 	2 errors occurred:
	// 		unexpected EOF
	// 		io: read/write on closed pipe
	// &errors.MultiError{errs:[]error{&errors.errorString{s:"EOF"}, &errors.MultiError{errs:[]error{&errors.errorString{s:"unexpected EOF"}, &errors.errorString{s:"io: read/write on closed pipe"}}}}}
}
