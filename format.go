package errors

import (
	"fmt"
	"io"
	"strings"
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

func goSyntax(errs []error) string {
	items := make([]string, len(errs))
	for i, err := range errs {
		items[i] = fmt.Sprintf("%#v", err)
	}

	return "[]error{" + strings.Join(items, ", ") + "}"
}

func indent(text string) string {
	return strings.ReplaceAll(text, "\n", "\n\t")
}
