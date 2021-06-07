package ext

import (
	"context"

	"github.com/flyteorg/flytectl/pkg/filters"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

func (a *AdminFetcherExtClient) ListProjects(ctx context.Context, filter filters.Filters) (*admin.Projects, error) {
	transformFilters, err := filters.BuildProjectListRequest(filter)
	if err != nil {
		return nil, err
	}
	e, err := a.AdminServiceClient().ListProjects(ctx, transformFilters)
	if err != nil {
		return nil, err
	}
	return e, nil
}
