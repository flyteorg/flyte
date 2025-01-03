package impl

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

const state = "state"

type NamedEntityMetrics struct {
	Scope promutils.Scope
}

type NamedEntityManager struct {
	db      repoInterfaces.Repository
	config  runtimeInterfaces.Configuration
	metrics NamedEntityMetrics
}

func (m *NamedEntityManager) UpdateNamedEntity(ctx context.Context, request *admin.NamedEntityUpdateRequest) (
	*admin.NamedEntityUpdateResponse, error) {
	if err := validation.ValidateNamedEntityUpdateRequest(request); err != nil {
		logger.Debugf(ctx, "invalid request [%+v]: %v", request, err)
		return nil, err
	}
	ctx = contextutils.WithProjectDomain(ctx, request.GetId().GetProject(), request.GetId().GetDomain())

	// Ensure entity exists before trying to update it
	_, err := util.GetNamedEntity(ctx, m.db, request.GetResourceType(), request.GetId())
	if err != nil {
		return nil, err
	}

	metadataModel := transformers.CreateNamedEntityModel(request)
	err = m.db.NamedEntityRepo().Update(ctx, metadataModel)
	if err != nil {
		logger.Debugf(ctx, "Failed to update named_entity for [%+v] with err %v", request.GetId(), err)
		return nil, err
	}
	return &admin.NamedEntityUpdateResponse{}, nil
}

func (m *NamedEntityManager) GetNamedEntity(ctx context.Context, request *admin.NamedEntityGetRequest) (
	*admin.NamedEntity, error) {
	if err := validation.ValidateNamedEntityGetRequest(request); err != nil {
		logger.Debugf(ctx, "invalid request [%+v]: %v", request, err)
		return nil, err
	}
	ctx = contextutils.WithProjectDomain(ctx, request.GetId().GetProject(), request.GetId().GetDomain())
	return util.GetNamedEntity(ctx, m.db, request.GetResourceType(), request.GetId())
}

func (m *NamedEntityManager) getQueryFilters(requestFilters string) ([]common.InlineFilter, error) {
	filters := make([]common.InlineFilter, 0)
	if len(requestFilters) == 0 {
		return filters, nil
	}
	additionalFilters, err := util.ParseFilters(requestFilters, common.NamedEntity)
	if err != nil {
		return nil, err
	}
	for _, filter := range additionalFilters {
		if strings.Contains(filter.GetField(), state) {
			filterWithDefaultValue, err := common.NewWithDefaultValueFilter(
				strconv.Itoa(int(admin.NamedEntityState_NAMED_ENTITY_ACTIVE)), filter)
			if err != nil {
				return nil, err
			}
			filters = append(filters, filterWithDefaultValue)
		} else {
			filters = append(filters, filter)
		}
	}
	return filters, nil
}

func (m *NamedEntityManager) ListNamedEntities(ctx context.Context, request *admin.NamedEntityListRequest) (
	*admin.NamedEntityList, error) {
	if err := validation.ValidateNamedEntityListRequest(request); err != nil {
		logger.Debugf(ctx, "invalid request [%+v]: %v", request, err)
		return nil, err
	}
	ctx = contextutils.WithProjectDomain(ctx, request.GetProject(), request.GetDomain())

	if len(request.GetFilters()) == 0 {
		// Add implicit filter to exclude system generated workflows
		request.Filters = fmt.Sprintf("not_like(name,%s)", ".flytegen%")
	}
	// HACK: In order to filter by state (if requested) - we need to amend the filter to use COALESCE
	// e.g. eq(state, 1) becomes 'WHERE (COALESCE(state, 0) = '1')' since not every NamedEntity necessarily
	// has an entry, and therefore the default state value '0' (active), should be assumed.
	filters, err := m.getQueryFilters(request.GetFilters())
	if err != nil {
		return nil, err
	}

	sortParameter, err := common.NewSortParameter(request.GetSortBy(), models.NamedEntityColumns)
	if err != nil {
		return nil, err
	}

	offset, err := validation.ValidateToken(request.GetToken())
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"invalid pagination token %s for ListNamedEntities", request.GetToken())
	}
	listInput := repoInterfaces.ListNamedEntityInput{
		ListResourceInput: repoInterfaces.ListResourceInput{
			Limit:         int(request.GetLimit()),
			Offset:        offset,
			InlineFilters: filters,
			SortParameter: sortParameter,
		},
		Project:      request.GetProject(),
		Domain:       request.GetDomain(),
		ResourceType: request.GetResourceType(),
	}

	output, err := m.db.NamedEntityRepo().List(ctx, listInput)
	if err != nil {
		logger.Debugf(ctx, "Failed to list named entities of type: %s with project: %s, domain: %s. Returned error was: %v",
			request.GetResourceType(), request.GetProject(), request.GetDomain(), err)
		return nil, err
	}

	var token string
	if len(output.Entities) == int(request.GetLimit()) {
		token = strconv.Itoa(offset + len(output.Entities))
	}
	entities := transformers.FromNamedEntityModels(output.Entities)
	return &admin.NamedEntityList{
		Entities: entities,
		Token:    token,
	}, nil

}

func NewNamedEntityManager(
	db repoInterfaces.Repository,
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
