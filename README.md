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

- **Drop-in replacement** for the standard `errors` package – swap the import, keep every call site.
- **Fields** – key-value context attached to an error, matched by `Is`.
- **Causes** – previous errors attached with `WithCause`, all visible to `Is` and `As`.
- **Stack traces** captured where the error is raised, printed with, or walked with `Frames`.
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
var ErrNotFound = errors.Sentinel("not found")  // no stack trace, captured when first derived from

err := errors.New("boom")                       // *Error with a stack trace
err = errors.Newf("user %d missing", 42)
err = errors.Wrap(io.EOF)                       // an *Error with a stack trace is returned unchanged
```

Adding context keeps everything the error already carries:

```go
err := ErrNotFound.WithField("id", 42)   // the stack trace is captured here
err = errors.Newf("load user: %w", err)  // fields, cause and stack are inherited
err = errors.Wrap(fmt.Errorf("handler: %w", err))

err.Fields()                                    // map[id:42]
errors.Is(err, ErrNotFound.WithField("id", 42)) // true
```

Fields and causes:

```go
err := ErrNotFound.WithField("id", 42)
err = err.WithFields(map[string]any{"table": "users", "attempt": 3})
err = err.WithCause(io.ErrUnexpectedEOF)

err = err.WithCause(io.ErrClosedPipe)

err.Fields()                        // map[attempt:3 id:42 table:users], a copy
err.Causes()                        // [unexpected EOF, io: read/write on closed pipe], a copy
errors.Is(err, io.ErrUnexpectedEOF) // true, every cause is in the chain
```

Sentinels match through fields, so declare them once and derive from them:

```go
errors.Is(err, ErrNotFound)                     // true
errors.Is(err, ErrNotFound.WithField("id", 42)) // true, fields are matched too
errors.Is(err, ErrNotFound.WithField("id", 7))  // false
errors.Is(err, errors.New("not found"))         // false, a different error with the same text
```

Check the error before wrapping it, never after:

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
errors.Join(nil, nil)        // nil
errors.Join(errA, errB)      // *MultiError, "2 errors occurred:\n\terrA\n\terrB"
errs.Len()                   // 2
errors.Is(errs, errA)        // true, every collected error is inspected
```

Printing, the message alone with `%v`, or the fields, the stack trace and the cause with `%+v`:

```go
fmt.Printf("%+v", err)
// process failed
//	process=abc
//	main.Process
//		/app/main.go:14
// caused by: connection refused
```

A multi error prints every collected error the same way:

```go
fmt.Printf("%+v", errs)
// 2 errors occurred:
//	process failed
//		process=abc
//		main.Process
//			/app/main.go:14
//	connection refused
```

Walking the captured stack:

```go
for frame := range err.Frames() {
	fmt.Println(frame.Function, frame.File, frame.Line)
}
```

Full reference: [pkg.go.dev][link-go-dev-reference].

## Conventions

**Standard library:** `Unwrap`, `Is`, `As`, `AsType` and `ErrUnsupported` are re-exported unchanged. `New` returns an
`*Error` instead of an `error`. `Join` returns an `error` whose dynamic type is `*MultiError`, and with two or more
errors its message is a numbered summary rather than the errors joined by newlines.

**Unwrapping:** `Error` and `MultiError` both implement `Unwrap() []error`, so `errors.Unwrap` returns `nil` for them.
Use `Is` and `As` to inspect the chain. An `Error` unwraps to its underlying error followed by its causes in the order
they were attached, so attaching a cause never hides the wrapped error nor an earlier cause. `WithCause(nil)` attaches
nothing.

**Equality:** Two `*Error` values match under `Is` when the underlying error of the target is found in the chain of
the underlying error of the inspected error, and every field of the target is present in the inspected error with the
same value. That chain never enters a cause, so fields only scope the errors they were added to or derived from; a
sentinel attached with `WithCause` is still found, but only with its own fields. Only the target itself is inspected,
never its cause nor the errors it wraps. Field values of different types never match, uncomparable values fall back
to `reflect.DeepEqual`, and an underlying error of an uncomparable type never matches.

**Typed nil:** `Wrap` returns `*Error` so that fields can be chained onto it. `Wrap(nil)` therefore returns a typed nil
pointer, which is not equal to `nil` once stored in an `error`. Chaining `WithField`, `WithFields` or `WithCause` on it
is safe and yields `nil` again, but the result must not be returned as an `error`.

**Formatting:** `%s`, `%q`, `%x` and `%v` print the message and honor width, precision, and flags. `%+v` prints the
underlying error with `%+v`, so a wrapped `MultiError` shows its members in full, then the fields sorted by key, the
stack trace innermost call first, and every cause formatted with `%+v` as well. `%#v` prints the error in Go syntax. `MultiError` prints its summary the same way, and with `%+v` every collected error is
formatted with `%+v` in turn.

**Stack traces:** `New`, `Newf` and `Wrap` capture up to 32 program counters above the caller; when the stack is deeper
the outermost calls are dropped, `Truncated` reports it and `%+v` ends the trace with `...`. `StackTrace` returns a copy,
and `Frames` may yield more frames than program counters when calls were inlined. `Sentinel` captures nothing, so
a package-level error does not point at package initialization; instead the first `Wrap`, `Newf`, `WithField`,
`WithFields` or `WithCause` applied to a stackless error captures the stack there, where the error is raised. `With*`
methods otherwise keep the stack of the error they derive from. When `Wrap` or `Newf` with `%w` reach an `*Error` by
unwrapping one error at a time, they inherit its fields, cause, and stack instead of capturing a new one, so context
can be added at every layer without losing anything. An error wrapping several errors at once, such as a
`MultiError`, is never entered, so nothing is inherited from one of its members.

**Multi error:** `Add` skips `nil` errors, typed `nil` values, and the collection itself, `Unwrap` returns a copy
safe to modify, and all methods are safe for concurrent use. A `nil` `*MultiError` reads as an empty collection, but
`Add` panics on it. A single collected error reports its message unchanged. `ErrorOrNil` is the only way to obtain a
`nil` `error` from a collection, so return it rather than the `*MultiError` itself.

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
