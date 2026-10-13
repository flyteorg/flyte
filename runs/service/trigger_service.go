package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"connectrpc.com/connect/v2"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	commonpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	taskpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	triggerpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/trigger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/trigger/triggerconnect"
	"github.com/flyteorg/flyte/v2/runs/repository/impl"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
	"github.com/flyteorg/flyte/v2/runs/repository/transformers"
	"github.com/flyteorg/flyte/v2/runs/schedule"
)

type triggerService struct {
	triggerconnect.UnimplementedTriggerServiceHandler
	db interfaces.Repository
}

func NewTriggerService(repo interfaces.Repository) triggerconnect.TriggerServiceHandler {
	return &triggerService{db: repo}
}

var _ triggerconnect.TriggerServiceHandler = (*triggerService)(nil)

func (s *triggerService) DeployTrigger(
	ctx context.Context,
	req *triggerpb.DeployTriggerRequest,
) (*triggerpb.DeployTriggerResponse, error) {
	request := req

	if err := request.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err.Error()).WithCause(err)
	}
	if err := validateCronExpression(request.GetName().GetName(), request.GetAutomationSpec()); err != nil {
		return nil, err
	}

	id := &commonpb.TriggerIdentifier{
		Name:     request.GetName(),
		Revision: request.GetRevision(),
	}
	triggerModel, err := transformers.NewTriggerModel(ctx, id, request.GetSpec(), request.GetAutomationSpec())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err.Error()).WithCause(err)
	}

	saved, err := s.db.TriggerRepo().SaveTrigger(ctx, triggerModel, request.GetRevision())
	if err != nil {
		logger.Errorf(ctx, "DeployTrigger failed: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err.Error()).WithCause(err)
	}

	details, err := transformers.TriggerModelToTriggerDetails(ctx, saved)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err.Error()).WithCause(err)
	}
	return &triggerpb.DeployTriggerResponse{Trigger: details}, nil
}

func (s *triggerService) GetTriggerDetails(
	ctx context.Context,
	req *triggerpb.GetTriggerDetailsRequest,
) (*triggerpb.GetTriggerDetailsResponse, error) {
	n := req.GetName()
	m, err := s.db.TriggerRepo().GetTrigger(ctx, transformers.ToTriggerKey(n))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, err.Error()).WithCause(err)
		}
		return nil, connect.NewError(connect.CodeInternal, err.Error()).WithCause(err)
	}
	details, err := transformers.TriggerModelToTriggerDetails(ctx, m)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err.Error()).WithCause(err)
	}
	return &triggerpb.GetTriggerDetailsResponse{Trigger: details}, nil
}

func (s *triggerService) GetTriggerRevisionDetails(
	ctx context.Context,
	req *triggerpb.GetTriggerRevisionDetailsRequest,
) (*triggerpb.GetTriggerRevisionDetailsResponse, error) {
	id := req.GetId()
	n := id.GetName()
	m, err := s.db.TriggerRepo().GetTriggerRevision(ctx, n.GetProject(), n.GetDomain(), n.GetTaskName(), n.GetName(), id.GetRevision())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, err.Error()).WithCause(err)
		}
		return nil, connect.NewError(connect.CodeInternal, err.Error()).WithCause(err)
	}
	details, err := transformers.TriggerRevisionModelToTriggerDetails(ctx, m)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err.Error()).WithCause(err)
	}
	return &triggerpb.GetTriggerRevisionDetailsResponse{Trigger: details}, nil
}

func (s *triggerService) ListTriggers(
	ctx context.Context,
	req *triggerpb.ListTriggersRequest,
) (*triggerpb.ListTriggersResponse, error) {
	request := req

	listInput, err := impl.NewListResourceInputFromProto(request.GetRequest(), models.TriggerColumns)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err.Error()).WithCause(err)
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

	ms, err := s.db.TriggerRepo().ListTriggers(ctx, listInput)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err.Error()).WithCause(err)
	}

	triggers, err := transformers.TriggerModelsToTriggers(ctx, ms)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err.Error()).WithCause(err)
	}

	var token string
	if len(triggers) > 0 && len(triggers) == listInput.Limit {
		token = fmt.Sprintf("%d", listInput.Offset+listInput.Limit)
	}
	return &triggerpb.ListTriggersResponse{Triggers: triggers, Token: token}, nil
}

func (s *triggerService) GetTriggerRevisionHistory(
	ctx context.Context,
	req *triggerpb.GetTriggerRevisionHistoryRequest,
) (*triggerpb.GetTriggerRevisionHistoryResponse, error) {
	request := req
	n := request.GetName()

	listInput, err := impl.NewListResourceInputFromProto(request.GetRequest(), models.TriggerRevisionColumns)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err.Error()).WithCause(err)
	}

	ms, err := s.db.TriggerRepo().ListTriggerRevisions(ctx,
		n.GetProject(), n.GetDomain(), n.GetTaskName(), n.GetName(), listInput)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err.Error()).WithCause(err)
	}

	revisions, err := transformers.TriggerRevisionModelsToTriggerRevisions(ctx, ms)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err.Error()).WithCause(err)
	}

	var token string
	if len(revisions) > 0 && len(revisions) == listInput.Limit {
		token = fmt.Sprintf("%d", listInput.Offset+listInput.Limit)
	}
	return &triggerpb.GetTriggerRevisionHistoryResponse{Triggers: revisions, Token: token}, nil
}

func (s *triggerService) UpdateTriggers(
	ctx context.Context,
	req *triggerpb.UpdateTriggersRequest,
) (*triggerpb.UpdateTriggersResponse, error) {
	request := req
	keys := namesToKeys(request.GetNames())
	if err := s.db.TriggerRepo().UpdateTriggers(ctx, keys, request.GetActive()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err.Error()).WithCause(err)
	}
	return &triggerpb.UpdateTriggersResponse{}, nil
}

func (s *triggerService) DeleteTriggers(
	ctx context.Context,
	req *triggerpb.DeleteTriggersRequest,
) (*triggerpb.DeleteTriggersResponse, error) {
	request := req
	keys := namesToKeys(request.GetNames())
	if err := s.db.TriggerRepo().DeleteTriggers(ctx, keys); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err.Error()).WithCause(err)
	}
	return &triggerpb.DeleteTriggersResponse{}, nil
}

// validateCronExpression validates the cron expression in a TriggerAutomationSpec, if present.
func validateCronExpression(triggerName string, spec *taskpb.TriggerAutomationSpec) error {
	if _, err := schedule.ParseCron(spec.GetSchedule()); err != nil {
		return connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("trigger %q has invalid cron expression: %w", triggerName, err).Error())
	}
	return nil
}

func namesToKeys(names []*commonpb.TriggerName) []interfaces.TriggerNameKey {
	keys := make([]interfaces.TriggerNameKey, 0, len(names))
	for _, n := range names {
		keys = append(keys, interfaces.NewTriggerNameKey(n.GetProject(), n.GetDomain(), n.GetTaskName(), n.GetName()))
	}
	return keys
}
