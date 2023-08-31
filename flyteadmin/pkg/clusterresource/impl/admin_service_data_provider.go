package impl

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"

	"github.com/flyteorg/flyteadmin/pkg/clusterresource/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

// Implementation of an interfaces.FlyteAdminDataProvider which fetches data using a flyteadmin service client
type serviceAdminProvider struct {
	adminClient service.AdminServiceClient
}

func (p serviceAdminProvider) GetClusterResourceAttributes(ctx context.Context, project, domain string) (*admin.ClusterResourceAttributes, error) {
	resource, err := p.adminClient.GetProjectDomainAttributes(ctx, &admin.ProjectDomainAttributesGetRequest{
		Project:      project,
		Domain:       domain,
		ResourceType: admin.MatchableResource_CLUSTER_RESOURCE,
	})
	if err != nil {
		return nil, err
	}
	if resource != nil && resource.Attributes != nil && resource.Attributes.MatchingAttributes != nil &&
		resource.Attributes.MatchingAttributes.GetClusterResourceAttributes() != nil {
		return resource.Attributes.MatchingAttributes.GetClusterResourceAttributes(), nil
	}
	return nil, NewMissingEntityError("cluster resource attributes")
}

var activeProjectsFilter = fmt.Sprintf("ne(state,%d)", admin.Project_ARCHIVED)

var descCreatedAtSortParam = admin.Sort{
	Direction: admin.Sort_DESCENDING,
	Key:       "created_at",
}

var descCreatedAtSortDBParam, _ = common.NewSortParameter(&descCreatedAtSortParam, models.ProjectColumns)

func (p serviceAdminProvider) GetProjects(ctx context.Context) (*admin.Projects, error) {
	projects := make([]*admin.Project, 0)
	listReq := &admin.ProjectListRequest{
		Limit:   100,
		Filters: activeProjectsFilter,
		// Prefer to sync projects most newly created to ensure their resources get created first when other resources exist.
		SortBy: &descCreatedAtSortParam,
	}

	// Iterate through all pages of projects
	for {
		projectResp, err := p.adminClient.ListProjects(ctx, listReq)
		if err != nil {
			return nil, err
		}
		projects = append(projects, projectResp.Projects...)
		if len(projectResp.Token) == 0 {
			break
		}
		listReq.Token = projectResp.Token
	}
	return &admin.Projects{
		Projects: projects,
	}, nil
}

func NewAdminServiceDataProvider(
	adminClient service.AdminServiceClient) interfaces.FlyteAdminDataProvider {
	return &serviceAdminProvider{
		adminClient: adminClient,
	}
}
