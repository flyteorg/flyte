package impl

import (
	"context"
	"strconv"

	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/util"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/validation"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/interfaces"
	repoInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/transformers"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

type ProjectManager struct {
	db     repoInterfaces.Repository
	config runtimeInterfaces.Configuration
}

var alphabeticalSortParam, _ = common.NewSortParameter(&admin.Sort{
	Direction: admin.Sort_ASCENDING,
	Key:       "identifier",
}, models.ProjectColumns)

func (m *ProjectManager) CreateProject(ctx context.Context, request admin.ProjectRegisterRequest) (
	*admin.ProjectRegisterResponse, error) {
	if err := validation.ValidateProjectRegisterRequest(request); err != nil {
		return nil, err
	}
	projectModel := transformers.CreateProjectModel(request.Project)
	err := m.db.ProjectRepo().Create(ctx, projectModel)
	if err != nil {
		return nil, err
	}

	return &admin.ProjectRegisterResponse{}, nil
}

func (m *ProjectManager) ListProjects(ctx context.Context, request admin.ProjectListRequest) (*admin.Projects, error) {
	spec := util.FilterSpec{
		RequestFilters: request.Filters,
	}
	filters, err := util.GetDbFilters(spec, common.Project)
	if err != nil {
		return nil, err
	}

	sortParameter, err := common.NewSortParameter(request.SortBy, models.ProjectColumns)
	if err != nil {
		return nil, err
	}
	if sortParameter == nil {
		sortParameter = alphabeticalSortParam
	}

	offset, err := validation.ValidateToken(request.Token)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"invalid pagination token %s for ListProjects", request.Token)
	}
	// And finally, query the database
	listProjectsInput := repoInterfaces.ListResourceInput{
		Limit:         int(request.Limit),
		Offset:        offset,
		InlineFilters: filters,
		SortParameter: sortParameter,
	}
	projectModels, err := m.db.ProjectRepo().List(ctx, listProjectsInput)
	if err != nil {
		return nil, err
	}
	projects := transformers.FromProjectModels(projectModels, m.GetDomains(ctx, admin.GetDomainRequest{}).Domains)

	var token string
	if len(projects) == int(request.Limit) {
		token = strconv.Itoa(offset + len(projects))
	}

	return &admin.Projects{
		Projects: projects,
		Token:    token,
	}, nil
}

func (m *ProjectManager) UpdateProject(ctx context.Context, projectUpdate admin.Project) (*admin.ProjectUpdateResponse, error) {
	var response admin.ProjectUpdateResponse
	projectRepo := m.db.ProjectRepo()

	// Fetch the existing project if exists. If not, return err and do not update.
	_, err := projectRepo.Get(ctx, projectUpdate.Id)
	if err != nil {
		return nil, err
	}

	// Run validation on the request and return err if validation does not succeed.
	if err := validation.ValidateProject(projectUpdate); err != nil {
		return nil, err
	}

	// Transform the provided project into a model and apply to the DB.
	projectUpdateModel := transformers.CreateProjectModel(&projectUpdate)
	err = projectRepo.UpdateProject(ctx, projectUpdateModel)

	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (m *ProjectManager) GetProject(ctx context.Context, request admin.ProjectGetRequest) (*admin.Project, error) {
	if err := validation.ValidateProjectGetRequest(request); err != nil {
		return nil, err
	}
	projectModel, err := m.db.ProjectRepo().Get(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	projectResponse := transformers.FromProjectModel(projectModel, m.GetDomains(ctx, admin.GetDomainRequest{}).Domains)

	return &projectResponse, nil
}

func (m *ProjectManager) GetDomains(ctx context.Context, request admin.GetDomainRequest) *admin.GetDomainsResponse {
	configDomains := m.config.ApplicationConfiguration().GetDomainsConfig()
	var domains = make([]*admin.Domain, len(*configDomains))
	for index, configDomain := range *configDomains {
		domains[index] = &admin.Domain{
			Id:   configDomain.ID,
			Name: configDomain.Name,
		}
	}
	return &admin.GetDomainsResponse{
		Domains: domains,
	}
}

func NewProjectManager(db repoInterfaces.Repository, config runtimeInterfaces.Configuration) interfaces.ProjectInterface {
	return &ProjectManager{
		db:     db,
		config: config,
	}
}
