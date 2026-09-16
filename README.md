# Errors

[![Latest Stable Version][ico-release]][link-release]
[![Build Status][ico-workflow]][link-workflow]
[![Coverage Status][ico-coverage]][link-coverage]
[![Go Dev Reference][ico-go-dev-reference]][link-go-dev-reference]
[![Software License][ico-license]][link-licence]

**Multi error**: Concurrent safe representation of a list of errors as a single error.

**Data error**: Additional context for error using data fields and cause (previous) error.


## Installation

```bash
go get github.com/gravitton/errors
```

## Drop-in replacement

This package is a drop-in replacement for the standard library `errors` package. It re-exports `Unwrap`, `Is`, `As`, and `AsType` unchanged, so you can swap the import and gain `Error` and `MultiError` without changing any existing call sites. 

`New` returns an `*Error` instead of an `error`. `Error` and `MultiError` both implement `Unwrap() []error`, so `errors.Unwrap` returns nil for them; use `Is` and `As` to inspect the chain.

```diff
- "errors"
+ "github.com/gravitton/errors"
```

## Usage

```go
import (
	"github.com/gravitton/errors"
)

func Process() error {
	if err := subProcess(); err != nil {
		return errors.Wrap(err).WithField("process", "abc")
	}

	return errors.Newf("this should not happen %s", "again")
}
```

```go
import (
	"sync"

	"github.com/gravitton/errors"
)

func Process() error {
	errs := errors.NewMulti()

	errs.Add(process(1), process(2))

	return errs.ErrorOrNil()
}

func ProcessConcurrent() error {
	errs := errors.NewMulti()
	wg := sync.WaitGroup{}

	for i := range 10 {
		wg.Go(func() {
			if err := process(i); err != nil {
				errs.Add(errors.Wrap(err).WithField("process", i))
			}
		})
	}

	wg.Wait()

	return errs.ErrorOrNil()
}
```

Print the message alone with `%v`, or the fields, the cause and the stack trace with `%+v`:

```go
fmt.Printf("%+v", err)
// process failed
//	process=abc
// caused by: connection refused
//	main.Process
//		/app/main.go:14
```

Walk the captured stack yourself with `Frames`:

```go
for frame := range err.Frames() {
	fmt.Println(frame.Function, frame.File, frame.Line)
}
```


## Credits

- [Tomáš Novotný](https://github.com/tomas-novotny)
- [All Contributors][link-contributors]


## License

The MIT License (MIT). Please see [License File][link-licence] for more information.


[ico-license]:              https://img.shields.io/github/license/gravitton/errors.svg?style=flat-square&colorB=blue
[ico-workflow]:             https://img.shields.io/github/actions/workflow/status/gravitton/errors/main.yml?branch=main&style=flat-square
[ico-release]:              https://img.shields.io/github/v/release/gravitton/errors?style=flat-square&colorB=blue
[ico-go-dev-reference]:     https://img.shields.io/badge/go.dev-reference-blue?style=flat-square
[ico-coverage]:             https://img.shields.io/coverallsCoverage/github/gravitton/errors?style=flat-square

[link-author]:              https://github.com/gravitton
[link-release]:             https://github.com/gravitton/errors/releases
[link-contributors]:        https://github.com/gravitton/errors/contributors
[link-licence]:             ./LICENSE.md
[link-changelog]:           ./CHANGELOG.md
[link-workflow]:            https://github.com/gravitton/errors/actions
[link-go-dev-reference]:    https://pkg.go.dev/github.com/gravitton/errors
[link-coverage]:            https://coveralls.io/github/gravitton/errors
