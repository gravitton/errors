// Package errors provides an extended error type with structured key-value
// fields, stack traces, and cause chaining, along with a thread-safe
// multi-error container. It is a drop-in superset of the standard library
// errors package: Unwrap, Is, and As are re-exported so callers only need to
// import this package.
package errors
