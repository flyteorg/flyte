package impl

import (
	"context"

	"github.com/lyft/flytestdlib/contextutils"

	"github.com/lyft/flyteadmin/pkg/manager/impl/validation"
	"github.com/lyft/flyteadmin/pkg/repositories/transformers"

	"github.com/lyft/flyteadmin/pkg/manager/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories"
	runtimeInterfaces "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

type ProjectDomainManager struct {
	db     repositories.RepositoryInterface
	config runtimeInterfaces.Configuration
}

func (m *ProjectDomainManager) UpdateProjectDomain(
	ctx context.Context, request admin.ProjectDomainAttributesUpdateRequest) (
	*admin.ProjectDomainAttributesUpdateResponse, error) {
	if err := validation.ValidateProjectDomainAttributesUpdateRequest(request); err != nil {
		return nil, err
	}
	ctx = contextutils.WithProjectDomain(ctx, request.Attributes.Project, request.Attributes.Domain)

	model, err := transformers.ToProjectDomainModel(*request.Attributes)
	if err != nil {
		return nil, err
	}
	err = m.db.ProjectDomainRepo().CreateOrUpdate(ctx, model)
	if err != nil {
		return nil, err
	}

	return &admin.ProjectDomainAttributesUpdateResponse{}, nil
}

func NewProjectDomainManager(
	db repositories.RepositoryInterface, config runtimeInterfaces.Configuration) interfaces.ProjectDomainInterface {
	return &ProjectDomainManager{
		db:     db,
		config: config,
	}
}
