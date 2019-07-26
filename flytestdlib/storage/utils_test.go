package storage

import (
	"os"
	"syscall"
	"testing"

	"github.com/graymeta/stow"
	flyteerrors "github.com/lyft/flytestdlib/errors"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestIsNotFound(t *testing.T) {
	sysError := &os.PathError{Err: syscall.ENOENT}
	assert.True(t, IsNotFound(sysError))
	flyteError := errors.Wrap(sysError, "Wrapping \"system not found\" error")
	assert.True(t, IsNotFound(flyteError))
	secondLevelError := errors.Wrap(flyteError, "Higher level error")
	assert.True(t, IsNotFound(secondLevelError))

	// more for stow errors
	stowNotFoundError := stow.ErrNotFound
	assert.True(t, IsNotFound(stowNotFoundError))
	flyteError = errors.Wrap(stowNotFoundError, "Wrapping stow.ErrNotFound")
	assert.True(t, IsNotFound(flyteError))
	secondLevelError = errors.Wrap(flyteError, "Higher level error wrapper of the stow.ErrNotFound error")
	assert.True(t, IsNotFound(secondLevelError))
}

func TestIsExceedsLimit(t *testing.T) {
	sysError := &os.PathError{Err: syscall.ENOENT}
	exceedsLimitError := flyteerrors.Wrapf(ErrExceedsLimit, sysError, "An error wrapped in ErrExceedsLimits")
	failedToWriteCacheError := flyteerrors.Wrapf(ErrFailedToWriteCache, sysError, "An error wrapped in ErrFailedToWriteCache")

	assert.True(t, IsExceedsLimit(exceedsLimitError))
	assert.False(t, IsExceedsLimit(failedToWriteCacheError))
	assert.False(t, IsExceedsLimit(sysError))
}

func TestIsFailedWriteToCache(t *testing.T) {
	sysError := &os.PathError{Err: syscall.ENOENT}
	exceedsLimitError := flyteerrors.Wrapf(ErrExceedsLimit, sysError, "An error wrapped in ErrExceedsLimits")
	failedToWriteCacheError := flyteerrors.Wrapf(ErrFailedToWriteCache, sysError, "An error wrapped in ErrFailedToWriteCache")

	assert.False(t, IsFailedWriteToCache(exceedsLimitError))
	assert.True(t, IsFailedWriteToCache(failedToWriteCacheError))
	assert.False(t, IsFailedWriteToCache(sysError))
}
