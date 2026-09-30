# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).


## [Unreleased](https://github.com/gravitton/errors/compare/v1.3.0...main)
### Added
- `Sentinel` creates a plain error without a stack trace, for package-level errors
- `ErrUnsupported` re-exported from the standard library
- `Error.GoString` prints a keyed Go literal for `%#v`, without the stack trace
- `Error.LogValue` logs the message, the fields and the cause as a `slog` group, without the stack trace
- `MultiError.LogValue` logs the members as a `slog` group keyed by their index
- `MultiError.Format` prints every member with `%+v` for `%+v`
- Examples of creating, wrapping, fields, causes, collections, formatting and logging

### Changed
- `Error.Unwrap` returns `[]error` with the underlying error and the cause, so `errors.Is` and `errors.As` see both, and `errors.Unwrap` returns nil for an `*Error` (**breaking**)
- `Error.Is` matches a target `*Error` by its underlying error instead of by message and fields, so copies made by `With*` match the error they came from (**breaking**)
- `Error.WithCause` joins a new cause with the attached ones into one flat `MultiError` instead of replacing them (**breaking**)
- `Error.WithCause` returns the error unchanged for an empty `MultiError`, as it does for `nil`
- `Error.WithField`, `WithFields` and `WithCause` panic on a nil receiver instead of returning nil (**breaking**)
- `Error.Format` prints for `%+v` the underlying error, the fields, the stack trace and then every cause, all with `%+v` and indented one tab
- `Error.Format` honors width, precision and flags for `%s`, `%v` and `%q`
- `Error.StackTrace` returns a copy
- `Wrap` and `Newf` reuse the stack trace of the first `*Error` in the `Unwrap() error` chain instead of capturing a new one
- `MultiError.Error` joins the messages by newlines like `errors.Join`, without the `N errors occurred:` header, and returns `<nil>` for a nil receiver (**breaking**)
- `MultiError.Unwrap` returns a copy
- `MultiError.GoString` prints a keyed Go literal like `Error.GoString`

### Removed
- `MultiError.Errors`, use `MultiError.Unwrap`, which now returns a copy (**breaking**)

### Fixed
- `Error.Frames` doc said outermost call first; frames are innermost first


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
