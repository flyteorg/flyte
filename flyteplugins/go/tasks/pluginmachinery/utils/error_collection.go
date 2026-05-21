package utils

import (
	"fmt"
	"strings"
)

type ErrorCollection struct {
	Errors []error
}

func (e ErrorCollection) Error() string {
	sb := strings.Builder{}
	for idx, err := range e.Errors {
		_, _ = fmt.Fprintf(&sb, "%v: %v\r\n", idx, err)
	}

	return sb.String()
}
