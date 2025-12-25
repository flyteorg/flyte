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

func (s *sortParameter) GetGormOrderExpr() string {
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

// GetSortByFieldsV2 converts proto sort fields to our SortParameter interfaces with validation
func GetSortByFieldsV2(request *common.ListRequest, allowedSortColumns sets.Set[string]) ([]interfaces.SortParameter, error) {
	if request == nil || request.GetSortByFields() == nil {
		return nil, nil
	}

	protoSortFields := request.GetSortByFields()
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
