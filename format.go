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
	if len(errs) == 0 {
		return "[]error(nil)"
	}

	items := make([]string, len(errs))
	for i, err := range errs {
		items[i] = fmt.Sprintf("%#v", err)
	}

	return "[]error{" + strings.Join(items, ", ") + "}"
}

func block(head string, children ...string) string {
	lines := hang(head)

	for _, child := range children {
		for _, line := range hang(child) {
			lines = append(lines, indent(line))
		}
	}

	return strings.Join(lines, "\n")
}

func hang(text string) []string {
	lines := strings.Split(text, "\n")
	for i, line := range lines[1:] {
		if !strings.HasPrefix(line, "\t") {
			lines[i+1] = indent(line)
		}
	}

	return lines
}

func indent(line string) string {
	if line == "" {
		return line
	}

	return "\t" + line
}
