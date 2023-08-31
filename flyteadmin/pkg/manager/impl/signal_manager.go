package impl

import (
	"context"
	"strconv"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/util"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/validation"
	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	repoInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"
)

type signalMetrics struct {
	Scope promutils.Scope
	Set   labeled.Counter
}

type SignalManager struct {
	db      repoInterfaces.Repository
	metrics signalMetrics
}

func getSignalContext(ctx context.Context, identifier *core.SignalIdentifier) context.Context {
	ctx = contextutils.WithProjectDomain(ctx, identifier.ExecutionId.Project, identifier.ExecutionId.Domain)
	ctx = contextutils.WithWorkflowID(ctx, identifier.ExecutionId.Name)
	return contextutils.WithSignalID(ctx, identifier.SignalId)
}

func (s *SignalManager) GetOrCreateSignal(ctx context.Context, request admin.SignalGetOrCreateRequest) (*admin.Signal, error) {
	if err := validation.ValidateSignalGetOrCreateRequest(ctx, request); err != nil {
		logger.Debugf(ctx, "invalid request [%+v]: %v", request, err)
		return nil, err
	}
	ctx = getSignalContext(ctx, request.Id)

	signalModel, err := transformers.CreateSignalModel(request.Id, request.Type, nil)
	if err != nil {
		logger.Errorf(ctx, "Failed to transform signal with id [%+v] and type [+%v] with err: %v", request.Id, request.Type, err)
		return nil, err
	}

	err = s.db.SignalRepo().GetOrCreate(ctx, &signalModel)
	if err != nil {
		return nil, err
	}

	signal, err := transformers.FromSignalModel(signalModel)
	if err != nil {
		logger.Errorf(ctx, "Failed to transform signal model [%+v] with err: %v", signalModel, err)
		return nil, err
	}

	return &signal, nil
}

func (s *SignalManager) ListSignals(ctx context.Context, request admin.SignalListRequest) (*admin.SignalList, error) {
	if err := validation.ValidateSignalListRequest(ctx, request); err != nil {
		logger.Debugf(ctx, "ListSignals request [%+v] is invalid: %v", request, err)
		return nil, err
	}
	ctx = getExecutionContext(ctx, request.WorkflowExecutionId)

	identifierFilters, err := util.GetWorkflowExecutionIdentifierFilters(ctx, *request.WorkflowExecutionId)
	if err != nil {
		return nil, err
	}

	filters, err := util.AddRequestFilters(request.Filters, common.Signal, identifierFilters)
	if err != nil {
		return nil, err
	}

	sortParameter, err := common.NewSortParameter(request.SortBy, models.SignalColumns)
	if err != nil {
		return nil, err
	}

	offset, err := validation.ValidateToken(request.Token)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"invalid pagination token %s for ListSignals", request.Token)
	}

	signalModelList, err := s.db.SignalRepo().List(ctx, repoInterfaces.ListResourceInput{
		InlineFilters: filters,
		Offset:        offset,
		Limit:         int(request.Limit),
		SortParameter: sortParameter,
	})
	if err != nil {
		logger.Debugf(ctx, "Failed to list signals with request [%+v] with err %v",
			request, err)
		return nil, err
	}

	signalList, err := transformers.FromSignalModels(signalModelList)
	if err != nil {
		logger.Debugf(ctx, "failed to transform signal models for request [%+v] with err: %v", request, err)
		return nil, err
	}
	var token string
	if len(signalList) == int(request.Limit) {
		token = strconv.Itoa(offset + len(signalList))
	}
	return &admin.SignalList{
		Signals: signalList,
		Token:   token,
	}, nil
}

func (s *SignalManager) SetSignal(ctx context.Context, request admin.SignalSetRequest) (*admin.SignalSetResponse, error) {
	if err := validation.ValidateSignalSetRequest(ctx, s.db, request); err != nil {
		return nil, err
	}
	ctx = getSignalContext(ctx, request.Id)

	signalModel, err := transformers.CreateSignalModel(request.Id, nil, request.Value)
	if err != nil {
		logger.Errorf(ctx, "Failed to transform signal with id [%+v] and value [+%v] with err: %v", request.Id, request.Value, err)
		return nil, err
	}

	err = s.db.SignalRepo().Update(ctx, signalModel.SignalKey, signalModel.Value)
	if err != nil {
		return nil, err
	}

	s.metrics.Set.Inc(ctx)
	return &admin.SignalSetResponse{}, nil
}

func NewSignalManager(
	db repoInterfaces.Repository,
	scope promutils.Scope) interfaces.SignalInterface {
	metrics := signalMetrics{
		Scope: scope,
		Set:   labeled.NewCounter("num_set", "count of set signals", scope),
	}

	return &SignalManager{
		db:      db,
		metrics: metrics,
	}
}
