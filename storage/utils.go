package storage

import (
	"os"

	errors2 "github.com/lyft/flytestdlib/errors"

	"github.com/graymeta/stow"
	"github.com/pkg/errors"
)

var (
	ErrExceedsLimit       errors2.ErrorCode = "LIMIT_EXCEEDED"
	ErrFailedToWriteCache errors2.ErrorCode = "CACHE_WRITE_FAILED"
)

// Gets a value indicating whether the underlying error is a Not Found error.
func IsNotFound(err error) bool {
	if root := errors.Cause(err); os.IsNotExist(root) {
		return true
	}

	if errors2.IsCausedByError(err, stow.ErrNotFound) {
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
	return errors2.IsCausedBy(err, ErrExceedsLimit)
}

func IsFailedWriteToCache(err error) bool {
	return errors2.IsCausedBy(err, ErrFailedToWriteCache)
}
