# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).


## [Unreleased](https://github.com/gravitton/errors/compare/v1.3.0...main)
### Added
- `Sentinel` creates an `*Error` without a stack trace, meant for package-level errors
- `ErrUnsupported` re-exported from the standard library
- `Error.GoString` and `MultiError.GoString` print the error and every contained error in Go syntax, used by `%#v`
- `MultiError.Format` implements `fmt.Formatter`: `%+v` prints every collected error with `%+v`
- Runnable examples of the error messages, fields, causes and collections

### Changed
- `Wrap`, `Newf`, `Error.Fields` and `%+v` look through single-error wrappers for inner `*Error` values: the first stack trace found is reused, fields are merged with the outermost value winning, and every cause is printed
- `Wrap`, `Error.WithField`, `Error.WithFields` and `Error.WithCause` give a stackless `*Error` the current stack trace
- `Error.WithCause` appends to the causes instead of replacing the cause (**breaking**)
- `Error.Unwrap` returns `[]error` holding the underlying error and the causes, so `errors.Unwrap` returns nil for an `*Error` (**breaking**)
- `Error.Is` matches when the target's underlying error is in the chain of the underlying error, instead of comparing the message and fields, so fields no longer scope a match and two `New` calls with the same text no longer match (**breaking**)
- `Error.Format` honors width, precision and flags for `%s`, `%q`, `%v` and `%x`, and `%+v` prints the causes with `%+v` after the stack trace, each indented one level
- `Error.StackTrace` and `MultiError.Unwrap` return copies
- `MultiError.Error` indents every message with a tab and reports `no errors` for an empty collection (**breaking**)
- Methods of `Error` and `MultiError` no longer handle nil receivers (**breaking**)

### Removed
- `Error.Frames`, resolve `Error.StackTrace` with `runtime.CallersFrames` instead (**breaking**)
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
