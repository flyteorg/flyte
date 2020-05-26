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
		sb.WriteString(fmt.Sprintf("%v: %v\r\n", idx, err))
	}

	return sb.String()
}
