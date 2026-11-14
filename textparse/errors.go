// Package textparse parses bounded color literals and single-state commands.
// It performs no I/O and does not control or discover devices.
package textparse

import (
	"fmt"
	"unicode/utf8"
)

// Error describes a rejected input. Start and End are a half-open UTF-8 byte
// range in the original input. Line and Column are one-based rune positions.
// No executable result is returned alongside an Error.
type Error struct {
	Code    string
	Field   string
	Start   int
	End     int
	Line    int
	Column  int
	Message string
	Hint    string
	Cause   error
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%s at %d:%d: %s", e.Code, e.Line, e.Column, e.Message)
}

// Unwrap preserves the underlying color or numeric conversion error.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func diagnostic(input string, t token, code, field, message, hint string, cause error) *Error {
	return &Error{Code: code, Field: field, Start: t.start, End: t.end,
		Line: 1, Column: utf8.RuneCountInString(input[:t.start]) + 1,
		Message: message, Hint: hint, Cause: cause}
}
