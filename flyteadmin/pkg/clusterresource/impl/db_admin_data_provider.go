package impl

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/clusterresource/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/clusterresource/plugin"
	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	managerInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/manager/interfaces"
	repositoryInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/transformers"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

const (
	stateColumn = "state"
)

// Implementation of an interfaces.FlyteAdminDataProvider which fetches data directly from the provided database connection.
type dbAdminProvider struct {
	db              repositoryInterfaces.Repository
	config          runtimeInterfaces.Configuration
	resourceManager managerInterfaces.ResourceInterface
}

func (p dbAdminProvider) GetClusterResourceAttributes(ctx context.Context, org, project, domain string) (*admin.ClusterResourceAttributes, error) {
	resource, err := p.resourceManager.GetResource(ctx, managerInterfaces.ResourceRequest{
		Org:          org,
		Project:      project,
		Domain:       domain,
		ResourceType: admin.MatchableResource_CLUSTER_RESOURCE,
	})
	if err != nil {
		return nil, err
	}
	if resource != nil && resource.Attributes != nil && resource.Attributes.GetClusterResourceAttributes() != nil {
		return resource.Attributes.GetClusterResourceAttributes(), nil
	}
	return nil, NewMissingEntityError("cluster resource attributes")
}

func (p dbAdminProvider) getDomains() []*admin.Domain {
	configDomains := p.config.ApplicationConfiguration().GetDomainsConfig()
	var domains = make([]*admin.Domain, len(*configDomains))
	for index, configDomain := range *configDomains {
		domains[index] = &admin.Domain{
			Id:   configDomain.ID,
			Name: configDomain.Name,
		}
	}
	return domains
}

func (p dbAdminProvider) getProjects(ctx context.Context, useActiveProjectsFilter bool) (projectsList *admin.Projects, err error) {
	var filter common.InlineFilter
	if useActiveProjectsFilter {
		filter, err = common.NewSingleValueFilter(common.Project, common.NotEqual, stateColumn, int32(admin.Project_ARCHIVED))
		if err != nil {
			return nil, err
		}
	} else {
		filter, err = common.NewSingleValueFilter(common.Project, common.Equal, stateColumn, int32(admin.Project_ARCHIVED))
		if err != nil {
			return nil, err
		}
	}

	projectModels, err := p.db.ProjectRepo().List(ctx, repositoryInterfaces.ListResourceInput{
		SortParameter: descCreatedAtSortDBParam,
		InlineFilters: []common.InlineFilter{filter},
	})
	if err != nil {
		return nil, err
	}
	projects := transformers.FromProjectModels(projectModels, p.getDomains())
	return &admin.Projects{
		Projects: projects,
	}, nil
}

func (p dbAdminProvider) GetProjects(ctx context.Context, clusterResourcePlugin plugin.ClusterResourcePlugin) (*admin.Projects, error) {
	return p.getProjects(ctx, getActiveProjects)
}

func (p dbAdminProvider) GetArchivedProjects(ctx context.Context, clusterResourcePlugin plugin.ClusterResourcePlugin) (*admin.Projects, error) {
	return p.getProjects(ctx, getArchivedProjects)
}

func NewDatabaseAdminDataProvider(db repositoryInterfaces.Repository, config runtimeInterfaces.Configuration, resourceManager managerInterfaces.ResourceInterface) interfaces.FlyteAdminDataProvider {
	return &dbAdminProvider{
		db:              db,
		config:          config,
		resourceManager: resourceManager,
	}
}
