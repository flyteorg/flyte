package storage

import (
	"os"
	"syscall"
	"testing"

	flyteerrors "github.com/flyteorg/flytestdlib/errors"
	"github.com/flyteorg/stow"
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

func TestMapStrings(t *testing.T) {
	t.Run("nothing", func(t *testing.T) {
		assert.Equal(t, []string{}, MapStrings(func(s string) string {
			return s
		}))
	})

	t.Run("one item", func(t *testing.T) {
		assert.Equal(t, []string{"item"}, MapStrings(func(s string) string {
			return s
		}, "item"))
	})

	t.Run("const", func(t *testing.T) {
		assert.Equal(t, []string{"something"}, MapStrings(func(s string) string {
			return "something"
		}, "item"))
	})

	t.Run("half string", func(t *testing.T) {
		assert.Equal(t, []string{"thing", "some"}, MapStrings(func(s string) string {
			return s[len(s)/2:]
		}, "something", "somesome"))
	})
}
