<div align="center" width="100%">

<a href="https://github.com/gravitton">
<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/gravitton/errors/refs/heads/main/docs/images/logo-dark.svg">
  <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/gravitton/errors/refs/heads/main/docs/images/logo-light.svg">
  <img alt="Gravitton errors" src="https://raw.githubusercontent.com/gravitton/errors/refs/heads/main/docs/images/logo-light.svg" width="300">
</picture>
</a>

[![Latest Stable Version][ico-release]][link-release]
[![Build Status][ico-workflow]][link-workflow]
[![Coverage Status][ico-coverage]][link-coverage]
[![Go Dev Reference][ico-go-dev-reference]][link-go-dev-reference]
[![Software License][ico-license]][link-licence]

Structured errors with fields, causes, and stack traces, plus a concurrent-safe multi error

<hr>

</div>


## Features

- **Drop-in replacement** for the standard `errors` package – swap the import, keep `New`, `Is`, `As` and `AsType`.
- **Fields** – key-value context added at any layer, merged along the chain.
- **Causes** – previous errors attached with `WithCause`, all visible to `Is` and `As`.
- **Stack traces** captured where the error is raised and printed with `%+v`.
- **Immutable** – every `With*` method returns a new error.
- **Multi error** – concurrent-safe collection of errors as a single `error`, with `Join` on top.

## Installation

```shell
go get github.com/gravitton/errors
```

## Usage

```diff
- "errors"
+ "github.com/gravitton/errors"
```

Creating and wrapping:

```go
var ErrNotFound = errors.Sentinel("not found")  // no stack trace, captured when derived or wrapped

err := errors.New("boom")                       // *Error with a stack trace
err = errors.Newf("user %d missing", 42)
err = errors.Wrap(io.EOF)                       // *Error with a stack trace around io.EOF
err = errors.Wrap(err)                          // an *Error with a stack trace is returned unchanged
```

Use `Sentinel`, not `New`, for package-level errors: `New` captures the stack where it is called, and at package
initialisation that stack points nowhere useful. For the same reason, don't derive package-level errors with
`WithField` or `WithCause`.

Adding context at every layer keeps what was added before:

```go
err := ErrNotFound.WithField("id", 42)  // the stack trace is captured here
err = errors.Newf("load user: %w", err) // the stack trace is reused
err = errors.Wrap(fmt.Errorf("handler: %w", err)).WithField("handler", "users")

err.Fields()                // map[handler:users id:42]
errors.Is(err, ErrNotFound) // true
```

Causes, the error that caused this one or one that happened while handling it:

```go
if err := load(); err != nil {
	return errors.New("could not load config").WithCause(err)
}
```

```go
if err := write(f); err != nil {
	if closeErr := f.Close(); closeErr != nil {
		return errors.Wrap(err).WithCause(closeErr)
	}

	return err
}
```

Fields are data about one occurrence, not part of the error's identity:

```go
errors.Is(err, ErrNotFound)                    // true
errors.Is(err, ErrNotFound.WithField("id", 7)) // true, fields are not compared
errors.Is(err, errors.New("not found"))        // false, a different error with the same text
```

Check the error before wrapping it, never after: `Wrap(nil)` returns a nil `*Error`, which is not `nil` once stored in
an `error`.

```go
func Process() error {
	if err := subProcess(); err != nil {
		return errors.Wrap(err).WithField("process", "abc")
	}

	return nil
}
```

Collecting errors, sequentially or concurrently:

```go
errs := errors.NewMulti()
errs.Add(process(1), process(2)) // nils are skipped

return errs.ErrorOrNil()         // nil when nothing was added
```

```go
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
```

```go
errors.Join(nil, nil)   // nil
errors.Join(errA, errB) // *MultiError, "2 errors occurred:\n\terrA\n\terrB"
```

Printing, the message alone with `%v`, or the fields, the stack trace and the causes with `%+v`:

```go
fmt.Printf("%+v", err)
// process failed
//	process=abc
//	main.Process
//		/app/main.go:14
// caused by: could not dial
//		main.dial
//			/app/main.go:27
//	caused by: connection refused
```

Each cause is indented one level, so the cause of a cause is told apart from its siblings.

`%+v` only reaches the fields and the stack trace when the outermost error is an `*Error`: `fmt.Errorf` does not
implement `fmt.Formatter`, so `fmt.Errorf("handler: %w", err)` prints the message alone. Wrap it, or use `Newf`.

Full reference: [pkg.go.dev][link-go-dev-reference].

## Conventions

- **Standard library:** `Unwrap`, `Is`, `As`, `AsType` and `ErrUnsupported` are re-exported. `errors.Unwrap` returns
  `nil` for an `*Error`, since it unwraps to its underlying error and its causes; use `Is` and `As`.
- **Wrapping:** `Wrap`, `Newf` and `Fields` look for inner `*Error` values through single-error wrappers only, never
  through causes or errors wrapping several errors, such as a `MultiError`. The first stack trace found is reused,
  otherwise up to 32 frames are captured.
- **Sentry:** `StackTrace` returns the program counters under the name Sentry looks for, so it picks them up.

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
