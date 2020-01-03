package impl

import (
	"context"

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

func NewProjectAttributesManager(db repositories.RepositoryInterface) interfaces.ProjectAttributesInterface {
	return &ProjectAttributesManager{
		db: db,
	}
}
