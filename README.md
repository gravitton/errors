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
- **Stack traces** captured where the error is raised and printed with `%+v`.
- **Fields** – key-value context attached to an error.
- **Cause** – the error that caused this one, visible to `Is` and `As`.
- **Identity** – every copy made by `With*` is still the error it came from for `errors.Is`.
- **Immutable** – every `With*` method returns a new error.
- **Structured logging** – errors are logged by `slog` as groups of their message, fields and cause.
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

### Creating

```go
var ErrNotFound = errors.Sentinel("not found") // a plain error without a stack trace

err := errors.New("boom")                      // *Error with the current stack trace
err = errors.Newf("user %d: %w", 42, ErrNotFound)
err = errors.Wrap(io.EOF)                      // *Error around io.EOF
err = errors.Wrap(err)                         // an *Error is returned unchanged
```

Use `Sentinel`, not `New`, for package-level errors: at package initialization, `New` captures a stack trace that
points nowhere useful. Wrap the sentinel where it is returned, which captures the stack trace there.

Check the error before wrapping it, never after: `Wrap(nil)` returns a nil `*Error`, which is not `nil` once stored in
an `error`. Reading a nil `*Error` or `*MultiError` is safe and prints `<nil>`, but `With*` and `Add` panic on one.

```go
if err := load(); err != nil {
	return errors.Wrap(err).WithField("file", name)
}

return nil
```

### Fields

```go
err := errors.Wrap(ErrNotFound).WithField("id", 42)
err = err.WithFields(map[string]any{"table": "users"})

err.Fields()                // map[id:42 table:users]
errors.Is(err, ErrNotFound) // true
```

Fields don't change identity: a copy made by `With*` is still the error it came from.

```go
err := errors.New("timeout")
errors.Is(err.WithField("attempt", 3), err) // true
```

Match a sentinel by the sentinel itself, not by a `Wrap` of it: an `*Error` target is only matched by an `*Error` in
the tree, never by a plain wrapper or a plain cause around the sentinel.

```go
err := fmt.Errorf("load: %w", ErrNotFound)
errors.Is(err, ErrNotFound)              // true
errors.Is(err, errors.Wrap(ErrNotFound)) // false
```

### Cause

The error that caused this one, or one that happened while handling it. Another cause is joined with the ones already
attached into one flat `MultiError`.

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

An empty collection attaches nothing, so a collection can be attached as it is:

```go
return errors.New("could not process batch").WithCause(errs)
```

### Wrapping

Wrapping an `*Error`, with `Wrap` or with `Newf` and `%w`, reuses its stack trace, which is closer to where the error
was raised. Only the `Unwrap() error` chain is searched: a collection, such as `Join`, gets the current stack trace.
The fields and cause of the wrapped error stay on it, reachable with `errors.As`:

```go
outer := errors.Newf("load user: %w", err)

outer.StackTrace()            // the stack trace of err
outer.Fields()                // map[]
errors.Is(outer, ErrNotFound) // true
```

### Collecting

A `MultiError` collects errors into a single `error`. It is safe for concurrent use, so goroutines can add to one
collection without extra locking.

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
errors.Join(errA, errB) // *MultiError, "errA\nerrB" as the standard errors.Join
```

### Printing

`%v` prints the message, `%+v` adds the fields, the stack trace and the cause, and `%#v` prints Go syntax without the
stack trace:

```go
fmt.Printf("%+v", err)
// process failed
//	process=abc
//	main.Process
//		/app/main.go:14
//	caused by: could not dial
//		main.dial
//			/app/main.go:27
//		caused by: connection refused
```

`%+v` applies to the underlying error and to every cause, so their details are printed too. An `*Error` wrapped with
`%w` is the exception: `fmt` prints only the wrapper's message, so its fields and cause are left out, reachable with
`errors.As`. A `MultiError` joins its members with newlines, as `errors.Join` does, with `%+v` applied to every member.

### Logging

```go
logger.Error("load failed", "err", err)
// level=ERROR msg="load failed" err.msg="not found" err.fields.id=42 err.fields.table=users

logger.Error("batch failed", "err", errors.Join(err, io.EOF))
// level=ERROR msg="batch failed" err.0.msg="not found" err.0.fields.id=42 err.0.fields.table=users err.1=EOF
```

The stack trace is left out of logs; it belongs to `%+v` and error reporting.

### Error reporting

`StackTrace() []uintptr` is the method Sentry looks for, so it picks up the stack trace; `Frames()` resolves it into
`runtime.Frame` values for other reporters. A reporter should walk the whole tree with `Unwrap` to collect the fields
of every `*Error` in it, and skip a stack trace equal to the one before, since wrapping errors share it.

Full reference: [pkg.go.dev][link-go-dev-reference].

## Differences from the standard library

- `errors.Unwrap` returns `nil` for every `*Error`, as it does for `errors.Join`: it only follows `Unwrap() error`,
  and `*Error` implements `Unwrap() []error` to expose both its underlying error and its cause. Use `Is` and `As`, or
  walk `Unwrap() []error`.
- `New` returns `*Error`, so `err := errors.New("x")` declares an `*Error` that can't be assigned a plain `error`
  later.
- Package-level `New` errors report package initialization as their stack trace; turn them into `Sentinel`s.
- `Join` returns a `*MultiError`, with the same message as the standard one.

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
