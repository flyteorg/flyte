package storage

import (
	"fmt"
	"os"
	"syscall"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	flyteerrors "github.com/flyteorg/flyte/v2/flytestdlib/errors"
	"github.com/flyteorg/stow"
)

// errorWrappings returns err bare and wrapped the ways stow and flytestdlib wrap errors: with pkg/errors, with %w and
// with both stacked in either order.
func errorWrappings(err error) map[string]error {
	return map[string]error{
		"bare":                err,
		"pkg/errors":          errors.Wrap(err, "wrapped"),
		"pkg/errors twice":    errors.Wrap(errors.Wrap(err, "wrapped"), "higher level"),
		"fmt %w":              fmt.Errorf("wrapped: %w", err),
		"pkg/errors over %w":  errors.Wrap(fmt.Errorf("wrapped: %w", err), "higher level"),
		"fmt %w over pkg/err": fmt.Errorf("higher level: %w", errors.Wrap(err, "wrapped")),
	}
}

func TestIsNotFound(t *testing.T) {
	notFound := map[string]error{
		"system not found":  &os.PathError{Err: syscall.ENOENT},
		"stow.ErrNotFound":  stow.ErrNotFound,
		"flyte wrapped":     flyteerrors.Wrapf("Code", stow.ErrNotFound, "wrapped"),
		"grpc status":       status.Error(codes.NotFound, "not found"),
		"os.ErrNotExist":    os.ErrNotExist,
		"syscall not found": syscall.ENOENT,
	}
	for errName, root := range notFound {
		for name, err := range errorWrappings(root) {
			if errName == "grpc status" && name != "bare" {
				// grpc status errors are only detected unwrapped.
				continue
			}

			t.Run(errName+"/"+name, func(t *testing.T) {
				assert.True(t, IsNotFound(err))
			})
		}
	}

	for name, err := range errorWrappings(&os.PathError{Err: syscall.EEXIST}) {
		t.Run("other error/"+name, func(t *testing.T) {
			assert.False(t, IsNotFound(err))
		})
	}

	assert.False(t, IsNotFound(nil))
}

func TestIsExists(t *testing.T) {
	exists := map[string]error{
		"system exists":  &os.PathError{Err: syscall.EEXIST},
		"os.ErrExist":    os.ErrExist,
		"syscall exists": syscall.EEXIST,
	}
	for errName, root := range exists {
		for name, err := range errorWrappings(root) {
			t.Run(errName+"/"+name, func(t *testing.T) {
				assert.True(t, IsExists(err))
			})
		}
	}

	for name, err := range errorWrappings(&os.PathError{Err: syscall.ENOENT}) {
		t.Run("other error/"+name, func(t *testing.T) {
			assert.False(t, IsExists(err))
		})
	}

	assert.False(t, IsExists(nil))
}

func TestIsExceedsLimit(t *testing.T) {
	sysError := &os.PathError{Err: syscall.ENOENT}
	exceedsLimitError := flyteerrors.Wrapf(ErrExceedsLimit, sysError, "An error wrapped in ErrExceedsLimits")
	failedToWriteCacheError := flyteerrors.Wrapf(ErrFailedToWriteCache, sysError, "An error wrapped in ErrFailedToWriteCache")

	assert.True(t, IsExceedsLimit(exceedsLimitError))
	for name, err := range errorWrappings(exceedsLimitError) {
		t.Run(name, func(t *testing.T) {
			assert.True(t, IsExceedsLimit(err))
		})
	}
	assert.False(t, IsExceedsLimit(failedToWriteCacheError))
	assert.False(t, IsExceedsLimit(sysError))
}

func TestIsFailedWriteToCache(t *testing.T) {
	sysError := &os.PathError{Err: syscall.ENOENT}
	exceedsLimitError := flyteerrors.Wrapf(ErrExceedsLimit, sysError, "An error wrapped in ErrExceedsLimits")
	failedToWriteCacheError := flyteerrors.Wrapf(ErrFailedToWriteCache, sysError, "An error wrapped in ErrFailedToWriteCache")

	assert.False(t, IsFailedWriteToCache(exceedsLimitError))
	assert.True(t, IsFailedWriteToCache(failedToWriteCacheError))
	for name, err := range errorWrappings(failedToWriteCacheError) {
		t.Run(name, func(t *testing.T) {
			assert.True(t, IsFailedWriteToCache(err))
		})
	}
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
