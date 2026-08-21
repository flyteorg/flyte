package service

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// validateRecovery checks a run's recovery relation before anything is persisted.
func (s *RunService) validateRecovery(ctx context.Context, runID *common.RunIdentifier, runSpec *task.RunSpec) error {
	relation := runSpec.GetRelation()
	source := relation.GetRelatedTo()
	// The recovery lookup reads actions by (project, domain, run_name, name) — the actions
	// table's primary key — and takes project/domain from related_to, not from the caller.
	// An unchecked cross-scope relation would therefore return another tenant's rows, and
	// their output URIs would be recorded as this run's recovered outputs. The SDK always
	// scopes related_to to the new run; this makes that a server-side guarantee.
	if source.GetProject() != runID.GetProject() || source.GetDomain() != runID.GetDomain() {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf(
			"cannot recover across scopes: source run %s/%s/%s is not in %s/%s",
			source.GetProject(), source.GetDomain(), source.GetName(),
			runID.GetProject(), runID.GetDomain()))
	}

	rootAction, err := s.repo.ActionRepo().GetAction(ctx, &common.ActionIdentifier{
		Run:  source,
		Name: RootActionName,
	})
	if err != nil {
		return connect.NewError(connect.CodeNotFound,
			fmt.Errorf("source run %s not found: %w", source.GetName(), err))
	}
	// If the source run is not in terminal phase, we cannot recover from it
	if !IsTerminalPhase(common.ActionPhase(rootAction.Phase)) {
		return connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf(
			"cannot recover from run %s: it is still in phase %s",
			source.GetName(), common.ActionPhase(rootAction.Phase)))
	}

	return nil
}

// LookupAction reports one action of a prior run for the enqueue-time recovery decision.
// A missing action is found = false, not an error; errors mean the lookup itself failed.
func (s *RunService) LookupAction(
	ctx context.Context,
	req *connect.Request[workflow.LookupActionRequest],
) (*connect.Response[workflow.LookupActionResponse], error) {
	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	actionID := req.Msg.GetActionId()
	action, err := s.repo.ActionRepo().GetAction(ctx, actionID)
	if err != nil {
		if errors.Is(err, interfaces.ErrActionNotFound) {
			return connect.NewResponse(&workflow.LookupActionResponse{Found: false}), nil
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := &workflow.LookupActionResponse{
		Found:       true,
		Phase:       common.ActionPhase(action.Phase),
		Attempts:    action.Attempts,
		CacheStatus: action.CacheStatus,
	}

	outputURI, err := s.resolveOutputURI(ctx, actionID, action)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	resp.OutputUri = outputURI

	return connect.NewResponse(resp), nil
}

// resolveOutputURI returns where an action's outputs live, or "" when it has none. A trace
// records the URI on its RunInfo; every other action type records it on its last attempt.
func (s *RunService) resolveOutputURI(
	ctx context.Context,
	actionID *common.ActionIdentifier,
	action *models.Action,
) (string, error) {
	if workflow.ActionType(action.ActionType) == workflow.ActionType_ACTION_TYPE_TRACE {
		if len(action.DetailedInfo) == 0 {
			return "", nil
		}
		info := &workflow.RunInfo{}
		if err := proto.Unmarshal(action.DetailedInfo, info); err != nil {
			return "", err
		}
		return info.GetOutputsUri(), nil
	}

	attempts, err := s.getAttempts(ctx, actionID)
	if err != nil {
		return "", err
	}
	if len(attempts) == 0 {
		return "", nil
	}
	return attempts[len(attempts)-1].GetOutputs().GetOutputUri(), nil
}
