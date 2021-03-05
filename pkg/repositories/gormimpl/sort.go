package gormimpl

import (
	"fmt"

	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	datacatalog "github.com/flyteorg/datacatalog/protos/gen"
)

const (
	sortQuery = "%s.%s %s"
)

// Container for the sort details
type sortParameter struct {
	sortKey   datacatalog.PaginationOptions_SortKey
	sortOrder datacatalog.PaginationOptions_SortOrder
}

// Generate the DBOrderExpression that GORM needs to order models
func (s *sortParameter) GetDBOrderExpression(tableName string) string {
	var sortOrderString string
	switch s.sortOrder {
	case datacatalog.PaginationOptions_ASCENDING:
		sortOrderString = "asc"
	case datacatalog.PaginationOptions_DESCENDING:
		fallthrough
	default:
		sortOrderString = "desc"
	}

	var sortKeyString string
	switch s.sortKey {
	case datacatalog.PaginationOptions_CREATION_TIME:
		fallthrough
	default:
		sortKeyString = "created_at"
	}
	return fmt.Sprintf(sortQuery, tableName, sortKeyString, sortOrderString)
}

// Create SortParameter for GORM
func NewGormSortParameter(sortKey datacatalog.PaginationOptions_SortKey, sortOrder datacatalog.PaginationOptions_SortOrder) models.SortParameter {
	return &sortParameter{sortKey: sortKey, sortOrder: sortOrder}
}
