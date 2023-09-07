package impl

import (
	"context"
	"strconv"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/util"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/validation"
	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	repoInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
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

func (m *ProjectManager) getDomains() []*admin.Domain {
	configDomains := m.config.ApplicationConfiguration().GetDomainsConfig()
	var domains = make([]*admin.Domain, len(*configDomains))
	for index, configDomain := range *configDomains {
		domains[index] = &admin.Domain{
			Id:   configDomain.ID,
			Name: configDomain.Name,
		}
	}
	return domains
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
	projects := transformers.FromProjectModels(projectModels, m.getDomains())

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

func NewProjectManager(db repoInterfaces.Repository, config runtimeInterfaces.Configuration) interfaces.ProjectInterface {
	return &ProjectManager{
		db:     db,
		config: config,
	}
}
