package impl

import (
	"context"

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

func NewWorkflowAttributesManager(db repositories.RepositoryInterface) interfaces.WorkflowAttributesInterface {
	return &WorkflowAttributesManager{
		db: db,
	}
}
