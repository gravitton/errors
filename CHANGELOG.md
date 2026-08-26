# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).


## [Unreleased](https://github.com/gravitton/errors/compare/v1.3.0...master)


## v1.3.0 (2026-08-26)(https://github.com/gravitton/errors/compare/v1.2.1...v1.3.0)
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


## v1.2.1 (2026-05-21)(https://github.com/gravitton/errors/compare/v1.2.0...v1.2.1)
### Fixed
- `DataError.Unwrap` now handles nil receiver safely, preventing a panic when a typed-nil `*DataError` is passed to `errors.Unwrap`, `errors.Is`, or `errors.As`


## v1.2.0 (2026-05-14)(https://github.com/gravitton/errors/compare/v1.1.1...v1.2.0)
### Added
- `AsType[T]` generic wrapper for `errors.AsType` (Go 1.26)
- `Join` constructor to build a `MultiError` from a variadic list of errors


## v1.1.1 (2026-04-23)(https://github.com/gravitton/errors/compare/v1.01.0...v1.1.1)
### Changed
- `Wrap` now accepts only `error` instead of `any`


## v1.1.0 (2026-04-22)(https://github.com/gravitton/errors/compare/v1.0.0...v1.1.0)
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
