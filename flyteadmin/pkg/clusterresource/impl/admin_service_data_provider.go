package impl

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/clusterresource/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/clusterresource/plugin"
	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

// Implementation of an interfaces.FlyteAdminDataProvider which fetches data using a flyteadmin service client
type serviceAdminProvider struct {
	adminClient service.AdminServiceClient
}

func (p serviceAdminProvider) GetClusterResourceAttributes(ctx context.Context, org, project, domain string) (*admin.ClusterResourceAttributes, error) {
	resource, err := p.adminClient.GetProjectDomainAttributes(ctx, &admin.ProjectDomainAttributesGetRequest{
		Org:          org,
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

var descUpdatedAtSortParam = admin.Sort{
	Direction: admin.Sort_DESCENDING,
	Key:       "updated_at",
}

var descCreatedAtSortDBParam, _ = common.NewSortParameter(&descUpdatedAtSortParam, models.ProjectColumns)

func (p serviceAdminProvider) getProjects(ctx context.Context, useActiveProjectsFilter bool, clusterResourcePlugin plugin.ClusterResourcePlugin) (*admin.Projects, error) {
	projects := make([]*admin.Project, 0)
	listReq := &admin.ProjectListRequest{
		Limit: 100,
		// Prefer to sync projects most newly updated to ensure their resources get modified first when other resources exist.
		SortBy: &descUpdatedAtSortParam,
	}
	var filter string
	var err error
	if useActiveProjectsFilter {
		filter, err = clusterResourcePlugin.GetProvisionProjectFilter(ctx)
	} else {
		filter, err = clusterResourcePlugin.GetDeprovisionProjectFilter(ctx)
	}
	if err != nil {
		return nil, err
	}
	listReq.Filters = filter
	logger.Debugf(ctx, "Fetching projects with filter [%s]", filter)

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

func (p serviceAdminProvider) GetProjects(ctx context.Context, clusterResourcePlugin plugin.ClusterResourcePlugin) (*admin.Projects, error) {
	return p.getProjects(ctx, getActiveProjects, clusterResourcePlugin)
}

func (p serviceAdminProvider) GetArchivedProjects(ctx context.Context, clusterResourcePlugin plugin.ClusterResourcePlugin) (*admin.Projects, error) {
	return p.getProjects(ctx, getArchivedProjects, clusterResourcePlugin)
}

func NewAdminServiceDataProvider(
	adminClient service.AdminServiceClient) interfaces.FlyteAdminDataProvider {
	return &serviceAdminProvider{
		adminClient: adminClient,
	}
}
