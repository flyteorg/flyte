package impl

import (
	"context"
	"strconv"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
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

type DescriptionEntityMetrics struct {
	Scope promutils.Scope
}

type DescriptionEntityManager struct {
	db      repoInterfaces.Repository
	config  runtimeInterfaces.Configuration
	metrics DescriptionEntityMetrics
}

func (d *DescriptionEntityManager) GetDescriptionEntity(ctx context.Context, request admin.ObjectGetRequest) (
	*admin.DescriptionEntity, error) {
	if err := validation.ValidateDescriptionEntityGetRequest(request); err != nil {
		logger.Errorf(ctx, "invalid request [%+v]: %v", request, err)
		return nil, err
	}
	ctx = contextutils.WithProjectDomain(ctx, request.Id.Project, request.Id.Domain)
	return util.GetDescriptionEntity(ctx, d.db, *request.Id)
}

func (d *DescriptionEntityManager) ListDescriptionEntity(ctx context.Context, request admin.DescriptionEntityListRequest) (*admin.DescriptionEntityList, error) {
	// Check required fields
	if err := validation.ValidateDescriptionEntityListRequest(request); err != nil {
		return nil, err
	}
	ctx = contextutils.WithProjectDomain(ctx, request.Id.Project, request.Id.Domain)

	if request.ResourceType == core.ResourceType_WORKFLOW {
		ctx = contextutils.WithWorkflowID(ctx, request.Id.Name)
	} else {
		ctx = contextutils.WithTaskID(ctx, request.Id.Name)
	}

	filters, err := util.GetDbFilters(util.FilterSpec{
		Project:        request.Id.Project,
		Domain:         request.Id.Domain,
		Name:           request.Id.Name,
		RequestFilters: request.Filters,
	}, common.ResourceTypeToEntity[request.ResourceType])
	if err != nil {
		logger.Error(ctx, "failed to get database filter")
		return nil, err
	}

	sortParameter, err := common.NewSortParameter(request.SortBy, models.DescriptionEntityColumns)
	if err != nil {
		return nil, err
	}

	offset, err := validation.ValidateToken(request.Token)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"invalid pagination token %s for ListWorkflows", request.Token)
	}
	listDescriptionEntitiesInput := repoInterfaces.ListResourceInput{
		Limit:         int(request.Limit),
		Offset:        offset,
		InlineFilters: filters,
		SortParameter: sortParameter,
	}
	output, err := d.db.DescriptionEntityRepo().List(ctx, listDescriptionEntitiesInput)
	if err != nil {
		logger.Debugf(ctx, "Failed to list workflows with [%+v] with err %v", request.Id, err)
		return nil, err
	}
	descriptionEntityList, err := transformers.FromDescriptionEntityModels(output.Entities)
	if err != nil {
		logger.Errorf(ctx,
			"Failed to transform workflow models [%+v] with err: %v", output.Entities, err)
		return nil, err
	}
	var token string
	if len(output.Entities) == int(request.Limit) {
		token = strconv.Itoa(offset + len(output.Entities))
	}
	return &admin.DescriptionEntityList{
		DescriptionEntities: descriptionEntityList,
		Token:               token,
	}, nil
}

func NewDescriptionEntityManager(
	db repoInterfaces.Repository,
	config runtimeInterfaces.Configuration,
	scope promutils.Scope) interfaces.DescriptionEntityInterface {

	metrics := DescriptionEntityMetrics{
		Scope: scope,
	}
	return &DescriptionEntityManager{
		db:      db,
		config:  config,
		metrics: metrics,
	}
}
