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

	// Parse token as offset
	var offset int
	if request.Token != "" {
		_, err := fmt.Sscanf(request.Token, "%d", &offset)
		if err != nil {
			return interfaces.ListResourceInput{}, fmt.Errorf("invalid token format: %s", request.Token)
		}
		if offset < 0 {
			return interfaces.ListResourceInput{}, fmt.Errorf("invalid offset: %d (must be non-negative)", offset)
		}
	}

	sortParameters, err := GetSortByFieldsV2(request, allowedColumns)
	if err != nil {
		return interfaces.ListResourceInput{}, err
	}

	combinedFilter, err := ConvertProtoFiltersToGormFilters(request.GetFilters())
	if err != nil {
		return interfaces.ListResourceInput{}, fmt.Errorf("failed to convert filters: %w", err)
	}

	return interfaces.ListResourceInput{
		Limit:          limit,
		Filter:         combinedFilter,
		Offset:         offset,
		SortParameters: sortParameters,
	}, nil
}
