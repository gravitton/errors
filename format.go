package errors

import (
	"fmt"
	"io"
	"strings"
)

// detailed is an error that can render itself with the fields, causes and
// stack traces it carries, as printed by the %+v verb.
type detailed interface {
	error
	fmt.GoStringer
	details() string
}

// format implements fmt.Formatter for a detailed error. The message is printed
// like a plain string, so %s, %q, %x and %v honour width, precision and flags.
// %+v prints the details and %#v prints the error in Go syntax.
func format(e detailed, s fmt.State, verb rune) {
	switch {
	case verb == 'v' && s.Flag('#'):
		io.WriteString(s, e.GoString())
	case verb == 'v' && s.Flag('+'):
		io.WriteString(s, e.details())
	default:
		fmt.Fprintf(s, fmt.FormatString(s, verb), e.Error())
	}
}

// indent prefixes every line of text with a tab.
func indent(text string) string {
	return "\t" + strings.ReplaceAll(text, "\n", "\n\t")
}
