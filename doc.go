// Package errors provides an error type with key-value fields, causes, and a
// stack trace, along with a multi-error collection safe for concurrent use. It
// is a drop-in superset of the standard library errors package: Unwrap, Is, As,
// AsType, and ErrUnsupported are re-exported and Join returns a MultiError, so
// callers only need to import this package.
package errors
