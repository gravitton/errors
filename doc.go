// Package errors is a drop-in superset of the standard library errors package: an [Error] with
// key-value fields, a cause and a stack trace, and a [MultiError] collecting several errors.
// [Unwrap], [Is], [As], [AsType] and [ErrUnsupported] are re-exported, so the import can be
// swapped.
//
// An Error is immutable: its With methods return the error changed and keep everything else,
// including the stack trace, which errors derived or wrapped from it share. Maps and slices
// returned are copies.
//
// A nil error, a nil *Error and a nil *MultiError are all a nil error: functions taking errors
// skip them, and reading a nil receiver never panics.
package errors
