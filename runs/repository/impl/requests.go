package impl

import (
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
)

// offsetFromToken parses the opaque page token as a non-negative integer offset. An empty
// token means the first page (offset 0). The RPC list handlers encode the next-page token
// as strconv.Itoa(offset+limit), matching the cloud service's pagination contract.
func offsetFromToken(token string) (int, error) {
	if token == "" {
		return 0, nil
	}
	offset, err := strconv.Atoi(token)
	if err != nil || offset < 0 {
		return 0, fmt.Errorf("invalid token: %q", token)
	}
	return offset, nil
}

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

	offset, err := offsetFromToken(request.GetToken())
	if err != nil {
		return interfaces.ListResourceInput{}, err
	}

	return interfaces.ListResourceInput{
		Limit:          limit,
		Offset:         offset,
		Filter:         combinedFilter,
		SortParameters: sortParameters,
	}, nil
}
