package storage

import (
	"fmt"
	"os"

	"github.com/graymeta/stow"
	"github.com/pkg/errors"
)

var ErrExceedsLimit = fmt.Errorf("limit exceeded")

// Gets a value indicating whether the underlying error is a Not Found error.
func IsNotFound(err error) bool {
	if root := errors.Cause(err); root == stow.ErrNotFound || os.IsNotExist(root) {
		return true
	}

	return false
}

// Gets a value indicating whether the underlying error is "already exists" error.
func IsExists(err error) bool {
	if root := errors.Cause(err); os.IsExist(root) {
		return true
	}

	return false
}

// Gets a value indicating whether the root cause of error is a "limit exceeded" error.
func IsExceedsLimit(err error) bool {
	return errors.Cause(err) == ErrExceedsLimit
}
