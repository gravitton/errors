package errors

import (
	"fmt"
	"io"
)

type detailed interface {
	error
	fmt.GoStringer
	details() string
}

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
