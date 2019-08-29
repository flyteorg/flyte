package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

func TestAlreadyExists(t *testing.T) {
	alreadyExistsErr := NewDataCatalogError(codes.AlreadyExists, "already exists")
	notFoundErr := NewDataCatalogError(codes.NotFound, "not found")
	assert.True(t, IsAlreadyExistsError(alreadyExistsErr))
	assert.False(t, IsAlreadyExistsError(notFoundErr))
}

func TestNotFoundErr(t *testing.T) {
	alreadyExistsErr := NewDataCatalogError(codes.AlreadyExists, "already exists")
	notFoundErr := NewDataCatalogError(codes.NotFound, "not found")
	assert.False(t, IsDoesNotExistError(alreadyExistsErr))
	assert.True(t, IsDoesNotExistError(notFoundErr))
}
