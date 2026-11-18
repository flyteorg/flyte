package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	commonpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	triggerpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/trigger"
	"github.com/flyteorg/flyte/v2/runs/repository/impl"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
	"github.com/flyteorg/flyte/v2/runs/repository/transformers"
	"github.com/flyteorg/flyte/v2/runs/schedule"
)

// Trigger provides transport-independent trigger operations.
type Trigger interface {
	DeployTrigger(context.Context, *triggerpb.DeployTriggerRequest) (*triggerpb.DeployTriggerResponse, error)
	GetTriggerDetails(context.Context, *triggerpb.GetTriggerDetailsRequest) (*triggerpb.GetTriggerDetailsResponse, error)
	GetTriggerRevisionDetails(context.Context, *triggerpb.GetTriggerRevisionDetailsRequest) (*triggerpb.GetTriggerRevisionDetailsResponse, error)
	ListTriggers(context.Context, *triggerpb.ListTriggersRequest) (*triggerpb.ListTriggersResponse, error)
	GetTriggerRevisionHistory(context.Context, *triggerpb.GetTriggerRevisionHistoryRequest) (*triggerpb.GetTriggerRevisionHistoryResponse, error)
	UpdateTriggers(context.Context, *triggerpb.UpdateTriggersRequest) (*triggerpb.UpdateTriggersResponse, error)
	DeleteTriggers(context.Context, *triggerpb.DeleteTriggersRequest) (*triggerpb.DeleteTriggersResponse, error)
}

// TriggerService manages triggers and their revisions.
type TriggerService struct {
	triggers interfaces.TriggerRepo
}

// NewTriggerService constructs a trigger service backed by the given repository.
func NewTriggerService(triggers interfaces.TriggerRepo) *TriggerService {
	return &TriggerService{triggers: triggers}
}

func (s *TriggerService) DeployTrigger(
	ctx context.Context,
	request *triggerpb.DeployTriggerRequest,
) (*triggerpb.DeployTriggerResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, ErrInvalidArgument.wrap(err)
	}
	if err := schedule.ValidateCronExpression(request.GetName().GetName(), request.GetAutomationSpec()); err != nil {
		return nil, ErrInvalidArgument.wrap(err)
	}

	id := &commonpb.TriggerIdentifier{
		Name:     request.GetName(),
		Revision: request.GetRevision(),
	}
	triggerModel, err := transformers.NewTriggerModel(ctx, id, request.GetSpec(), request.GetAutomationSpec())
	if err != nil {
		return nil, ErrInternal.wrap(err)
	}

	saved, err := s.triggers.SaveTrigger(ctx, triggerModel, request.GetRevision())
	if err != nil {
		logger.Errorf(ctx, "DeployTrigger failed: %v", err)
		return nil, ErrInternal.wrap(err)
	}

	details, err := transformers.TriggerModelToTriggerDetails(ctx, saved)
	if err != nil {
		return nil, ErrInternal.wrap(err)
	}
	return &triggerpb.DeployTriggerResponse{Trigger: details}, nil
}

func (s *TriggerService) GetTriggerDetails(
	ctx context.Context,
	request *triggerpb.GetTriggerDetailsRequest,
) (*triggerpb.GetTriggerDetailsResponse, error) {
	n := request.GetName()
	m, err := s.triggers.GetTrigger(ctx, transformers.ToTriggerKey(n))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound.wrap(err)
		}
		return nil, ErrInternal.wrap(err)
	}
	details, err := transformers.TriggerModelToTriggerDetails(ctx, m)
	if err != nil {
		return nil, ErrInternal.wrap(err)
	}
	return &triggerpb.GetTriggerDetailsResponse{Trigger: details}, nil
}

func (s *TriggerService) GetTriggerRevisionDetails(
	ctx context.Context,
	request *triggerpb.GetTriggerRevisionDetailsRequest,
) (*triggerpb.GetTriggerRevisionDetailsResponse, error) {
	id := request.GetId()
	n := id.GetName()
	m, err := s.triggers.GetTriggerRevision(ctx, n.GetProject(), n.GetDomain(), n.GetTaskName(), n.GetName(), id.GetRevision())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound.wrap(err)
		}
		return nil, ErrInternal.wrap(err)
	}
	details, err := transformers.TriggerRevisionModelToTriggerDetails(ctx, m)
	if err != nil {
		return nil, ErrInternal.wrap(err)
	}
	return &triggerpb.GetTriggerRevisionDetailsResponse{Trigger: details}, nil
}

func (s *TriggerService) ListTriggers(
	ctx context.Context,
	request *triggerpb.ListTriggersRequest,
) (*triggerpb.ListTriggersResponse, error) {
	listInput, err := impl.NewListResourceInputFromProto(request.GetRequest(), models.TriggerColumns)
	if err != nil {
		return nil, ErrInvalidArgument.wrap(err)
	}

	switch scope := request.GetScopeBy().(type) {
	case *triggerpb.ListTriggersRequest_ProjectId:
		listInput.ScopeByFilter = impl.NewProjectIdFilter(scope.ProjectId)
	case *triggerpb.ListTriggersRequest_TaskId:
		tid := scope.TaskId
		listInput.ScopeByFilter = impl.NewEqualFilter("project", tid.GetProject()).
			And(impl.NewEqualFilter("domain", tid.GetDomain())).
			And(impl.NewEqualFilter("task_name", tid.GetName())).
			And(impl.NewEqualFilter("task_version", tid.GetVersion()))
	case *triggerpb.ListTriggersRequest_TaskName:
		tn := scope.TaskName
		listInput.ScopeByFilter = impl.NewEqualFilter("project", tn.GetProject()).
			And(impl.NewEqualFilter("domain", tn.GetDomain())).
			And(impl.NewEqualFilter("task_name", tn.GetName()))
	}

	ms, err := s.triggers.ListTriggers(ctx, listInput)
	if err != nil {
		return nil, ErrInternal.wrap(err)
	}

	triggers, err := transformers.TriggerModelsToTriggers(ctx, ms)
	if err != nil {
		return nil, ErrInternal.wrap(err)
	}

	var token string
	if len(triggers) > 0 && len(triggers) == listInput.Limit {
		token = fmt.Sprintf("%d", listInput.Offset+listInput.Limit)
	}
	return &triggerpb.ListTriggersResponse{Triggers: triggers, Token: token}, nil
}

func (s *TriggerService) GetTriggerRevisionHistory(
	ctx context.Context,
	request *triggerpb.GetTriggerRevisionHistoryRequest,
) (*triggerpb.GetTriggerRevisionHistoryResponse, error) {
	n := request.GetName()

	listInput, err := impl.NewListResourceInputFromProto(request.GetRequest(), models.TriggerRevisionColumns)
	if err != nil {
		return nil, ErrInvalidArgument.wrap(err)
	}

	ms, err := s.triggers.ListTriggerRevisions(ctx,
		n.GetProject(), n.GetDomain(), n.GetTaskName(), n.GetName(), listInput)
	if err != nil {
		return nil, ErrInternal.wrap(err)
	}

	revisions, err := transformers.TriggerRevisionModelsToTriggerRevisions(ctx, ms)
	if err != nil {
		return nil, ErrInternal.wrap(err)
	}

	var token string
	if len(revisions) > 0 && len(revisions) == listInput.Limit {
		token = fmt.Sprintf("%d", listInput.Offset+listInput.Limit)
	}
	return &triggerpb.GetTriggerRevisionHistoryResponse{Triggers: revisions, Token: token}, nil
}

func (s *TriggerService) UpdateTriggers(
	ctx context.Context,
	request *triggerpb.UpdateTriggersRequest,
) (*triggerpb.UpdateTriggersResponse, error) {
	keys := namesToKeys(request.GetNames())
	if err := s.triggers.UpdateTriggers(ctx, keys, request.GetActive()); err != nil {
		return nil, ErrInternal.wrap(err)
	}
	return &triggerpb.UpdateTriggersResponse{}, nil
}

func (s *TriggerService) DeleteTriggers(
	ctx context.Context,
	request *triggerpb.DeleteTriggersRequest,
) (*triggerpb.DeleteTriggersResponse, error) {
	keys := namesToKeys(request.GetNames())
	if err := s.triggers.DeleteTriggers(ctx, keys); err != nil {
		return nil, ErrInternal.wrap(err)
	}
	return &triggerpb.DeleteTriggersResponse{}, nil
}

func namesToKeys(names []*commonpb.TriggerName) []interfaces.TriggerNameKey {
	keys := make([]interfaces.TriggerNameKey, 0, len(names))
	for _, n := range names {
		keys = append(keys, interfaces.NewTriggerNameKey(n.GetProject(), n.GetDomain(), n.GetTaskName(), n.GetName()))
	}
	return keys
}
