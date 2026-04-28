package impl

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
)

// sortParameter implements the SortParameter interface
type sortParameter struct {
	field string
	order interfaces.SortOrder
}

func (s *sortParameter) GetOrderExpr() string {
	orderStr := "DESC"
	if s.order == interfaces.SortOrderAscending {
		orderStr = "ASC"
	}
	return fmt.Sprintf("%s %s", s.field, orderStr)
}

// NewSortParameter creates a new sort parameter
func NewSortParameter(field string, order interfaces.SortOrder) interfaces.SortParameter {
	return &sortParameter{
		field: field,
		order: order,
	}
}

// GetSortByFieldsV2 converts proto sort fields to our SortParameter interfaces with validation.
// It prefers sort_by_fields; if empty, it falls back to the deprecated sort_by field.
func GetSortByFieldsV2(request *common.ListRequest, allowedSortColumns sets.Set[string]) ([]interfaces.SortParameter, error) {
	if request == nil {
		return nil, nil
	}

	protoSortFields := request.GetSortByFields()
	if len(protoSortFields) == 0 {
		// Fall back to deprecated sort_by field.
		if sortBy := request.GetSortBy(); sortBy != nil && sortBy.Key != "" {
			protoSortFields = []*common.Sort{sortBy}
		}
	}
	if len(protoSortFields) == 0 {
		return nil, nil
	}

	sortParams := make([]interfaces.SortParameter, 0, len(protoSortFields))

	for _, sortField := range protoSortFields {
		// Validate field is allowed
		if !allowedSortColumns.Has(sortField.Key) {
			return nil, fmt.Errorf("invalid sort field: %s", sortField.Key)
		}

		// Convert direction
		var order interfaces.SortOrder
		if sortField.Direction == common.Sort_ASCENDING {
			order = interfaces.SortOrderAscending
		} else {
			order = interfaces.SortOrderDescending
		}

		sortParams = append(sortParams, NewSortParameter(sortField.Key, order))
	}

	return sortParams, nil
}
