package impl

import (
	"context"

	"github.com/lyft/flytestdlib/logger"

	"github.com/lyft/flyteadmin/pkg/manager/impl/validation"
	"github.com/lyft/flyteadmin/pkg/repositories/transformers"

	"github.com/lyft/flyteadmin/pkg/manager/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

type ProjectAttributesManager struct {
	db repositories.RepositoryInterface
}

func (m *ProjectAttributesManager) UpdateProjectAttributes(
	ctx context.Context, request admin.ProjectAttributesUpdateRequest) (
	*admin.ProjectAttributesUpdateResponse, error) {
	var resource admin.MatchableResource
	var err error
	if resource, err = validation.ValidateProjectAttributesUpdateRequest(request); err != nil {
		return nil, err
	}

	model, err := transformers.ToProjectAttributesModel(*request.Attributes, resource)
	if err != nil {
		return nil, err
	}
	err = m.db.ProjectAttributesRepo().CreateOrUpdate(ctx, model)
	if err != nil {
		return nil, err
	}

	return &admin.ProjectAttributesUpdateResponse{}, nil
}

func (m *ProjectAttributesManager) GetProjectAttributes(ctx context.Context, request admin.ProjectAttributesGetRequest) (
	*admin.ProjectAttributesGetResponse, error) {
	if err := validation.ValidateProjectAttributesGetRequest(request); err != nil {
		return nil, err
	}
	projectAttributesModel, err := m.db.ProjectAttributesRepo().Get(ctx, request.Project, request.ResourceType.String())
	if err != nil {
		return nil, err
	}
	projectAttributes, err := transformers.FromProjectAttributesModel(projectAttributesModel)
	if err != nil {
		return nil, err
	}
	return &admin.ProjectAttributesGetResponse{
		Attributes: &projectAttributes,
	}, nil
}

func (m *ProjectAttributesManager) DeleteProjectAttributes(ctx context.Context,
	request admin.ProjectAttributesDeleteRequest) (*admin.ProjectAttributesDeleteResponse, error) {
	if err := validation.ValidateProjectAttributesDeleteRequest(request); err != nil {
		return nil, err
	}
	if err := m.db.ProjectAttributesRepo().Delete(ctx, request.Project, request.ResourceType.String()); err != nil {
		return nil, err
	}
	logger.Infof(ctx, "Deleted project attributes for: %s (%s)", request.Project, request.ResourceType.String())
	return &admin.ProjectAttributesDeleteResponse{}, nil
}

func NewProjectAttributesManager(db repositories.RepositoryInterface) interfaces.ProjectAttributesInterface {
	return &ProjectAttributesManager{
		db: db,
	}
}
