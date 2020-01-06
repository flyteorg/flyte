package impl

import (
	"context"

	"github.com/lyft/flytestdlib/logger"

	"github.com/lyft/flytestdlib/contextutils"

	"github.com/lyft/flyteadmin/pkg/manager/impl/validation"
	"github.com/lyft/flyteadmin/pkg/repositories/transformers"

	"github.com/lyft/flyteadmin/pkg/manager/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

type ProjectDomainAttributesManager struct {
	db repositories.RepositoryInterface
}

func (m *ProjectDomainAttributesManager) UpdateProjectDomainAttributes(
	ctx context.Context, request admin.ProjectDomainAttributesUpdateRequest) (
	*admin.ProjectDomainAttributesUpdateResponse, error) {
	var resource admin.MatchableResource
	var err error
	if resource, err = validation.ValidateProjectDomainAttributesUpdateRequest(request); err != nil {
		return nil, err
	}
	ctx = contextutils.WithProjectDomain(ctx, request.Attributes.Project, request.Attributes.Domain)

	model, err := transformers.ToProjectDomainAttributesModel(*request.Attributes, resource)
	if err != nil {
		return nil, err
	}
	err = m.db.ProjectDomainAttributesRepo().CreateOrUpdate(ctx, model)
	if err != nil {
		return nil, err
	}

	return &admin.ProjectDomainAttributesUpdateResponse{}, nil
}

func (m *ProjectDomainAttributesManager) GetProjectDomainAttributes(
	ctx context.Context, request admin.ProjectDomainAttributesGetRequest) (
	*admin.ProjectDomainAttributesGetResponse, error) {
	if err := validation.ValidateProjectDomainAttributesGetRequest(request); err != nil {
		return nil, err
	}
	projectAttributesModel, err := m.db.ProjectDomainAttributesRepo().Get(
		ctx, request.Project, request.Domain, request.ResourceType.String())
	if err != nil {
		return nil, err
	}
	projectAttributes, err := transformers.FromProjectDomainAttributesModel(projectAttributesModel)
	if err != nil {
		return nil, err
	}
	return &admin.ProjectDomainAttributesGetResponse{
		Attributes: &projectAttributes,
	}, nil
}

func (m *ProjectDomainAttributesManager) DeleteProjectDomainAttributes(ctx context.Context,
	request admin.ProjectDomainAttributesDeleteRequest) (*admin.ProjectDomainAttributesDeleteResponse, error) {
	if err := validation.ValidateProjectDomainAttributesDeleteRequest(request); err != nil {
		return nil, err
	}
	if err := m.db.ProjectDomainAttributesRepo().Delete(
		ctx, request.Project, request.Domain, request.ResourceType.String()); err != nil {
		return nil, err
	}
	logger.Infof(ctx, "Deleted project-domain attributes for: %s-%s (%s)", request.Project,
		request.Domain, request.ResourceType.String())
	return &admin.ProjectDomainAttributesDeleteResponse{}, nil
}

func NewProjectDomainAttributesManager(
	db repositories.RepositoryInterface) interfaces.ProjectDomainAttributesInterface {
	return &ProjectDomainAttributesManager{
		db: db,
	}
}
