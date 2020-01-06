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

type WorkflowAttributesManager struct {
	db repositories.RepositoryInterface
}

func (m *WorkflowAttributesManager) UpdateWorkflowAttributes(
	ctx context.Context, request admin.WorkflowAttributesUpdateRequest) (
	*admin.WorkflowAttributesUpdateResponse, error) {
	var resource admin.MatchableResource
	var err error
	if resource, err = validation.ValidateWorkflowAttributesUpdateRequest(request); err != nil {
		return nil, err
	}

	model, err := transformers.ToWorkflowAttributesModel(*request.Attributes, resource)
	if err != nil {
		return nil, err
	}
	err = m.db.WorkflowAttributesRepo().CreateOrUpdate(ctx, model)
	if err != nil {
		return nil, err
	}

	return &admin.WorkflowAttributesUpdateResponse{}, nil
}

func (m *WorkflowAttributesManager) GetWorkflowAttributes(
	ctx context.Context, request admin.WorkflowAttributesGetRequest) (
	*admin.WorkflowAttributesGetResponse, error) {
	if err := validation.ValidateWorkflowAttributesGetRequest(request); err != nil {
		return nil, err
	}
	projectAttributesModel, err := m.db.WorkflowAttributesRepo().Get(
		ctx, request.Project, request.Domain, request.Workflow, request.ResourceType.String())
	if err != nil {
		return nil, err
	}
	projectAttributes, err := transformers.FromWorkflowAttributesModel(projectAttributesModel)
	if err != nil {
		return nil, err
	}
	return &admin.WorkflowAttributesGetResponse{
		Attributes: &projectAttributes,
	}, nil
}

func (m *WorkflowAttributesManager) DeleteWorkflowAttributes(ctx context.Context,
	request admin.WorkflowAttributesDeleteRequest) (*admin.WorkflowAttributesDeleteResponse, error) {
	if err := validation.ValidateWorkflowAttributesDeleteRequest(request); err != nil {
		return nil, err
	}
	if err := m.db.WorkflowAttributesRepo().Delete(
		ctx, request.Project, request.Domain, request.Workflow, request.ResourceType.String()); err != nil {
		return nil, err
	}
	logger.Infof(ctx, "Deleted workflow attributes for: %s-%s-%s (%s)", request.Project,
		request.Domain, request.Workflow, request.ResourceType.String())
	return &admin.WorkflowAttributesDeleteResponse{}, nil
}

func NewWorkflowAttributesManager(db repositories.RepositoryInterface) interfaces.WorkflowAttributesInterface {
	return &WorkflowAttributesManager{
		db: db,
	}
}
