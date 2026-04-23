package impl

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
)

const (
	defaultLimit = 50
	maxLimit     = 200
)

// NewListResourceInputFromProto converts a proto ListRequest to ListResourceInput for querying DB
func NewListResourceInputFromProto(request *common.ListRequest, allowedColumns sets.Set[string]) (interfaces.ListResourceInput, error) {
	if request == nil {
		request = &common.ListRequest{}
	}

	// Validate and parse limit
	limit := int(request.GetLimit())
	if limit == 0 {
		limit = defaultLimit
	}
	if limit < 0 {
		return interfaces.ListResourceInput{}, fmt.Errorf("invalid limit: %d (must be non-negative)", limit)
	}
	if limit > maxLimit {
		return interfaces.ListResourceInput{}, fmt.Errorf("invalid limit: %d (exceeds maximum of %d)", limit, maxLimit)
	}

	sortParameters, err := GetSortByFieldsV2(request, allowedColumns)
	if err != nil {
		return interfaces.ListResourceInput{}, err
	}

	combinedFilter, err := ConvertProtoFilters(request.GetFilters(), allowedColumns)
	if err != nil {
		return interfaces.ListResourceInput{}, fmt.Errorf("failed to convert filters: %w", err)
	}

	return interfaces.ListResourceInput{
		Limit:          limit,
		Filter:         combinedFilter,
		CursorToken:    request.Token,
		SortParameters: sortParameters,
	}, nil
}
