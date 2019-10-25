package errors

import (
	"testing"

	"fmt"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestErrorHelpers(t *testing.T) {
	alreadyExistsErr := NewDataCatalogError(codes.AlreadyExists, "already exists")
	notFoundErr := NewDataCatalogError(codes.NotFound, "not found")

	t.Run("TestAlreadyExists", func(t *testing.T) {
		assert.True(t, IsAlreadyExistsError(alreadyExistsErr))
		assert.False(t, IsAlreadyExistsError(notFoundErr))
	})

	t.Run("TestNotFoundErr", func(t *testing.T) {
		assert.False(t, IsDoesNotExistError(alreadyExistsErr))
		assert.True(t, IsDoesNotExistError(notFoundErr))
	})

	t.Run("TestCollectErrs", func(t *testing.T) {
		collectedErr := NewCollectedErrors(codes.InvalidArgument, []error{alreadyExistsErr, notFoundErr})
		assert.EqualValues(t, status.Code(collectedErr), codes.InvalidArgument)
		assert.Equal(t, collectedErr.Error(), fmt.Sprintf("%s, %s", alreadyExistsErr.Error(), notFoundErr.Error()))
	})
}
