package service

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect/v2"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/project"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/project/projectconnect"
	"github.com/flyteorg/flyte/v2/runs/repository/impl"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
	"github.com/flyteorg/flyte/v2/runs/repository/transformers"
)

type ProjectService struct {
	projectRepo interfaces.ProjectRepo
	domains     []*project.Domain
}

func NewProjectService(projectRepo interfaces.ProjectRepo, domains []*project.Domain) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
		domains:     domains,
	}
}

func (s *ProjectService) CreateProject(
	ctx context.Context,
	req *project.CreateProjectRequest,
) (*project.CreateProjectResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err.Error()).WithCause(err)
	}

	projectProto := req.GetProject()
	if projectProto == nil || projectProto.GetId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, "project.id is required")
	}

	model, err := transformers.NewProjectModel(projectProto)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err.Error()).WithCause(err)
	}

	if err := s.projectRepo.CreateProject(ctx, model); err != nil {
		if errors.Is(err, interfaces.ErrProjectAlreadyExists) {
			logger.Warnf(ctx, "Project already exists, identifier: %s, name: %s", model.Identifier, model.Name)
			return nil, connect.NewError(connect.CodeAlreadyExists, err.Error()).WithCause(err)
		}
		return nil, connect.NewError(connect.CodeInternal, err.Error()).WithCause(err)
	}

	return &project.CreateProjectResponse{}, nil
}

func (s *ProjectService) UpdateProject(
	ctx context.Context,
	req *project.UpdateProjectRequest,
) (*project.UpdateProjectResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err.Error()).WithCause(err)
	}

	projectProto := req.GetProject()
	if projectProto == nil || projectProto.GetId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, "project.id is required")
	}

	model, err := transformers.NewProjectModel(projectProto)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err.Error()).WithCause(err)
	}

	if err := s.projectRepo.UpdateProject(ctx, model); err != nil {
		if errors.Is(err, interfaces.ErrProjectNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err.Error()).WithCause(err)
		}
		return nil, connect.NewError(connect.CodeInternal, err.Error()).WithCause(err)
	}

	return &project.UpdateProjectResponse{}, nil
}

func (s *ProjectService) GetProject(
	ctx context.Context,
	req *project.GetProjectRequest,
) (*project.GetProjectResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err.Error()).WithCause(err)
	}
	if req.GetId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, "id is required")
	}

	model, err := s.projectRepo.GetProject(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, interfaces.ErrProjectNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err.Error()).WithCause(err)
		}
		return nil, connect.NewError(connect.CodeInternal, err.Error()).WithCause(err)
	}

	projectProto, err := transformers.ProjectModelToProject(model, s.domains)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err.Error()).WithCause(err)
	}

	return &project.GetProjectResponse{
		Project: projectProto,
	}, nil
}

func (s *ProjectService) ListProjects(
	ctx context.Context,
	req *project.ListProjectsRequest,
) (*project.ListProjectsResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err.Error()).WithCause(err)
	}

	listInput, err := listResourceInputFromProjectListRequest(req)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err.Error()).WithCause(err)
	}

	// If no filters are supplied, hide archived (deleted) projects.
	if req.GetFilters() == "" {
		listInput.Filter = impl.NewNotEqualFilter("state", int32(project.ProjectState_PROJECT_STATE_ARCHIVED))
	}

	projects, err := s.projectRepo.ListProjects(ctx, listInput)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err.Error()).WithCause(err)
	}

	projectProtos := make([]*project.Project, 0, len(projects))
	for _, p := range projects {
		projectProto, convErr := transformers.ProjectModelToProject(p, s.domains)
		if convErr != nil {
			return nil, connect.NewError(connect.CodeInternal, convErr.Error()).WithCause(convErr)
		}
		projectProtos = append(projectProtos, projectProto)
	}

	var token string
	if listInput.Limit > 0 && len(projectProtos) == listInput.Limit {
		token = fmt.Sprintf("%d", listInput.Offset+listInput.Limit)
	}

	return &project.ListProjectsResponse{
		Projects: &project.Projects{
			Projects: projectProtos,
			Token:    token,
		},
	}, nil
}

func listResourceInputFromProjectListRequest(req *project.ListProjectsRequest) (interfaces.ListResourceInput, error) {
	limit := int(req.GetLimit())
	var offset int
	if req.GetToken() != "" {
		if _, err := fmt.Sscanf(req.GetToken(), "%d", &offset); err != nil {
			return interfaces.ListResourceInput{}, fmt.Errorf("invalid token format: %s", req.GetToken())
		}
		if offset < 0 {
			return interfaces.ListResourceInput{}, fmt.Errorf("invalid offset: %d (must be non-negative)", offset)
		}
	}

	listInput := interfaces.ListResourceInput{
		Limit:  limit,
		Offset: offset,
	}

	if sortField := req.GetSortBy().GetKey(); sortField != "" {
		if !models.ProjectColumns.Has(sortField) {
			return interfaces.ListResourceInput{}, fmt.Errorf("invalid sort field: %s", sortField)
		}

		sortOrder := interfaces.SortOrderDescending
		if req.GetSortBy().GetDirection() == common.Sort_ASCENDING {
			sortOrder = interfaces.SortOrderAscending
		}
		listInput.SortParameters = []interfaces.SortParameter{
			impl.NewSortParameter(sortField, sortOrder),
		}
	}

	if req.GetFilters() != "" {
		filter, err := impl.ParseStringFilters(req.GetFilters(), models.ProjectColumns)
		if err != nil {
			return interfaces.ListResourceInput{}, err
		}
		listInput.Filter = filter
	}

	return listInput, nil
}

var _ projectconnect.ProjectServiceHandler = (*ProjectService)(nil)
