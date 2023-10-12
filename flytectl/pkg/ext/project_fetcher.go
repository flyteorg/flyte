package ext

import (
	"context"
	"fmt"

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

func (a *AdminFetcherExtClient) GetProjectByID(ctx context.Context, projectID string) (*admin.Project, error) {
	if projectID == "" {
		return nil, fmt.Errorf("GetProjectByID: projectId is empty")
	}

	response, err := a.AdminServiceClient().ListProjects(ctx, &admin.ProjectListRequest{
		Limit:   1,
		Filters: fmt.Sprintf("eq(identifier,%s)", filters.EscapeValue(projectID)),
	})
	if err != nil {
		return nil, err
	}

	if len(response.Projects) == 0 {
		return nil, NewNotFoundError("project %s", projectID)
	}

	if len(response.Projects) > 1 {
		panic(fmt.Sprintf("unexpected number of projects in ListProjects response: %d - 0 or 1 expected", len(response.Projects)))
	}

	return response.Projects[0], nil
}
