# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).


## [Unreleased](https://github.com/gravitton/errors/compare/v1.3.0...main)
### Added
- `Sentinel` creates a plain error without a stack trace, for package-level errors
- `ErrUnsupported` re-exported from the standard library
- `Error.GoString`, used by `%#v`: a keyed Go literal of the fields that are set, without the stack trace
- `Error.LogValue` and `MultiError.LogValue` implement `slog.LogValuer`: an `*Error` is logged as a group of its message, a `fields` group and its cause, without the stack trace, and a `MultiError` as a group of its members keyed by their index
- `MultiError.Format` implements `fmt.Formatter`: `%+v` joins the members formatted with `%+v` by newlines
- Examples of creating, wrapping, fields, causes, collections and their `%+v`, `%#v` and `slog` output

### Changed
- `Wrap` and `Newf` reuse the stack trace of the first `*Error` in the wrapped error's `Unwrap() error` chain instead of capturing a new one
- `Error.Unwrap` returns `[]error` holding the underlying error and the cause, so both are visible to `errors.Is` and `errors.As`, and `errors.Unwrap` returns nil for an `*Error` (**breaking**)
- `Error.Is` matches a target `*Error` by its underlying error instead of by message and fields, so copies made by `With*` match the error they came from, and a target without an underlying error, such as the zero value, matches nothing (**breaking**)
- `Error.WithCause` joins another cause with the ones already attached into one flat `MultiError` instead of replacing it (**breaking**)
- `Error.Format` honors width, precision and flags for `%s`, `%v` and `%q`, and `%+v` prints the underlying error with `%+v`, then the fields, the stack trace and every cause, each member of a joined cause on its own `caused by:` line, formatted with `%+v`, each indented one tab
- `Error.StackTrace` returns a copy
- `MultiError.Error` joins the messages by newlines like `errors.Join`, without the `N errors occurred:` header (**breaking**)
- `MultiError.GoString` prints a keyed Go literal like `Error.GoString`
- Nil receivers: `Error.WithField`, `Error.WithFields`, `Error.WithCause` and `MultiError.Add` panic instead of returning nil or doing nothing, and `MultiError.Error` returns `<nil>` instead of an empty string, as `Error.Error` does (**breaking**)

### Removed
- `MultiError.Errors`, use `MultiError.Unwrap` instead (**breaking**)


## [v1.3.0](https://github.com/gravitton/errors/compare/v1.2.1...v1.3.0) (2026-08-26)
### Added
- `Error.Format` implements `fmt.Formatter`: `%+v` prints fields, cause and stack trace
- `Error.Frames` iterator resolving the captured stack trace into `runtime.Frame` values
- `MultiError.Len` and `MultiError.Errors` accessors
- `MultiError.Add` accepts a variadic list of errors

### Changed
- Require Go 1.27
- `DataError` renamed to `Error` (**breaking**)
- `Error.Fields` returns a copy, so the returned map can no longer mutate the error
- `Error.WithFields` no longer drops function values
- `Error.WithField`, `WithFields` and `WithCause` return nil for a nil receiver instead of panicking

### Fixed
- `Error.Is` no longer panics when a field holds an uncomparable value such as a slice or a map
- `Error.Error`, `Fields`, `StackTrace` and `Is` now handle a nil receiver safely


## [v1.2.1](https://github.com/gravitton/errors/compare/v1.2.0...v1.2.1) (2026-05-21)
### Fixed
- `DataError.Unwrap` now handles nil receiver safely, preventing a panic when a typed-nil `*DataError` is passed to `errors.Unwrap`, `errors.Is`, or `errors.As`


## [v1.2.0](https://github.com/gravitton/errors/compare/v1.1.1...v1.2.0) (2026-05-14)
### Added
- `AsType[T]` generic wrapper for `errors.AsType` (Go 1.26)
- `Join` constructor to build a `MultiError` from a variadic list of errors


## [v1.1.1](https://github.com/gravitton/errors/compare/v1.1.0...v1.1.1) (2026-04-23)
### Changed
- `Wrap` now accepts only `error` instead of `any`


## [v1.1.0](https://github.com/gravitton/errors/compare/v1.0.0...v1.1.0) (2026-04-22)
### Added
- `DataError` stack trace capture via `StackTrace()`

### Changed
- `MultiError.Append` renamed to `MultiError.Add`
- `MultiError` mutex upgraded to `sync.RWMutex` for improved read concurrency
- `MultiError.ErrorOrNil` and `Unwrap` now handle nil receiver safely
- `MultiError.Error` and `GoString` are now concurrency-safe


## v1.0.0 (2025-10-17)
### Added
- `DataError` with additional context fields and cause error
- `MultiError` concurrent safe multi error
