package ext

import (
	"errors"
	"fmt"
)

type NotFoundError struct {
	Target string
}

func (err *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found", err.Target)
}

func NewNotFoundError(targetFormat string, formatArgs ...any) *NotFoundError {
	target := fmt.Sprintf(targetFormat, formatArgs...)
	return &NotFoundError{target}
}

func IsNotFoundError(err error) bool {
	var notFoundErr *NotFoundError
	return errors.As(err, &notFoundErr)
}
