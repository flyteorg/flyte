package impl

import (
	"context"
	"strconv"

	"github.com/lyft/flytestdlib/contextutils"

	"github.com/lyft/flyteadmin/pkg/common"
	"github.com/lyft/flyteadmin/pkg/errors"
	"google.golang.org/grpc/codes"

	"github.com/lyft/flyteadmin/pkg/manager/impl/util"
	"github.com/lyft/flyteadmin/pkg/manager/impl/validation"
	"github.com/lyft/flyteadmin/pkg/manager/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories"
	repoInterfaces "github.com/lyft/flyteadmin/pkg/repositories/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories/transformers"
	runtimeInterfaces "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/promutils"
)

type NamedEntityMetrics struct {
	Scope promutils.Scope
}

type NamedEntityManager struct {
	db      repositories.RepositoryInterface
	config  runtimeInterfaces.Configuration
	metrics NamedEntityMetrics
}

func (m *NamedEntityManager) UpdateNamedEntity(ctx context.Context, request admin.NamedEntityUpdateRequest) (
	*admin.NamedEntityUpdateResponse, error) {
	if err := validation.ValidateNamedEntityUpdateRequest(request); err != nil {
		logger.Debugf(ctx, "invalid request [%+v]: %v", request, err)
		return nil, err
	}
	ctx = contextutils.WithProjectDomain(ctx, request.Id.Project, request.Id.Domain)

	// Ensure entity exists before trying to update it
	_, err := util.GetNamedEntity(ctx, m.db, request.ResourceType, *request.Id)
	if err != nil {
		return nil, err
	}

	metadataModel := transformers.CreateNamedEntityModel(&request)
	err = m.db.NamedEntityRepo().Update(ctx, metadataModel)
	if err != nil {
		logger.Debugf(ctx, "Failed to update named_entity for [%+v] with err %v", request.Id, err)
		return nil, err
	}
	return &admin.NamedEntityUpdateResponse{}, nil
}

func (m *NamedEntityManager) GetNamedEntity(ctx context.Context, request admin.NamedEntityGetRequest) (
	*admin.NamedEntity, error) {
	if err := validation.ValidateNamedEntityGetRequest(request); err != nil {
		logger.Debugf(ctx, "invalid request [%+v]: %v", request, err)
		return nil, err
	}
	ctx = contextutils.WithProjectDomain(ctx, request.Id.Project, request.Id.Domain)
	return util.GetNamedEntity(ctx, m.db, request.ResourceType, *request.Id)
}

func (m *NamedEntityManager) ListNamedEntities(ctx context.Context, request admin.NamedEntityListRequest) (
	*admin.NamedEntityList, error) {
	if err := validation.ValidateNamedEntityListRequest(request); err != nil {
		logger.Debugf(ctx, "invalid request [%+v]: %v", request, err)
		return nil, err
	}
	ctx = contextutils.WithProjectDomain(ctx, request.Project, request.Domain)

	filters, err := util.GetDbFilters(util.FilterSpec{
		Project: request.Project,
		Domain:  request.Domain,
	}, common.ResourceTypeToEntity[request.ResourceType])
	if err != nil {
		return nil, err
	}
	var sortParameter common.SortParameter
	if request.SortBy != nil {
		sortParameter, err = common.NewSortParameter(*request.SortBy)
		if err != nil {
			return nil, err
		}
	}
	offset, err := validation.ValidateToken(request.Token)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"invalid pagination token %s for ListNamedEntities", request.Token)
	}
	listInput := repoInterfaces.ListResourceInput{
		Limit:         int(request.Limit),
		Offset:        offset,
		InlineFilters: filters,
		SortParameter: sortParameter,
	}

	output, err := m.db.NamedEntityRepo().List(ctx, request.ResourceType, listInput)
	if err != nil {
		logger.Debugf(ctx, "Failed to list named entities of type: %s with project: %s, domain: %s. Returned error was: %v",
			request.ResourceType, request.Project, request.Domain, err)
		return nil, err
	}

	var token string
	if len(output.Entities) == int(request.Limit) {
		token = strconv.Itoa(offset + len(output.Entities))
	}
	entities := transformers.FromNamedEntityModels(output.Entities)
	return &admin.NamedEntityList{
		Entities: entities,
		Token:    token,
	}, nil

}

func NewNamedEntityManager(
	db repositories.RepositoryInterface,
	config runtimeInterfaces.Configuration,
	scope promutils.Scope) interfaces.NamedEntityInterface {

	metrics := NamedEntityMetrics{
		Scope: scope,
	}
	return &NamedEntityManager{
		db:      db,
		config:  config,
		metrics: metrics,
	}
}
