# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).


## [Unreleased](https://github.com/gravitton/errors/compare/v1.1.0...master)

## v1.1.0 (2026-04-22)(https://github.com/gravitton/errors/compare/v1.0.0...v1.1.0)
### Added
- `DataError` stack trace capture via `StackTrace()`
- `DataError.Is` for sentinel matching with optional field scoping
- `DataError.WithCause` for attaching a cause to the error chain
- `DataError.GoString` for debugging output

### Changed
- `MultiError.Append` renamed to `MultiError.Add`
- `MultiError` mutex upgraded to `sync.RWMutex` for improved read concurrency
- `MultiError.ErrorOrNil` and `Unwrap` now handle nil receiver safely
- `MultiError.Error` and `GoString` are now concurrency-safe

## v1.0.0 (2025-10-17)
### Added
- `DataError` with additional context fields and cause error
- `MultiError` concurrent safe multi error
