package gormimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
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
