package gormimpl

import (
	"testing"

	datacatalog "github.com/flyteorg/datacatalog/protos/gen"
	"github.com/stretchr/testify/assert"
)

func TestSortAsc(t *testing.T) {
	dbSortExpression := NewGormSortParameter(
		datacatalog.PaginationOptions_CREATION_TIME,
		datacatalog.PaginationOptions_ASCENDING).GetDBOrderExpression("artifacts")

	assert.Equal(t, dbSortExpression, "artifacts.created_at asc")
}

func TestSortDesc(t *testing.T) {
	dbSortExpression := NewGormSortParameter(
		datacatalog.PaginationOptions_CREATION_TIME,
		datacatalog.PaginationOptions_DESCENDING).GetDBOrderExpression("artifacts")

	assert.Equal(t, dbSortExpression, "artifacts.created_at desc")
}
