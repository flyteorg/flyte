package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	executorv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
	statek8s "github.com/flyteorg/flyte/v2/state/k8s"
)

// RunService implements the RunServiceHandler interface
type RunService struct {
	repo          interfaces.Repository
	queueClient   workflowconnect.QueueServiceClient
	storagePrefix string
	dataStore     *storage.DataStore
	stateClient   *statek8s.StateClient
}

// WatchGroups streams task groups (runs grouped by task) from the database.
func (s *RunService) WatchGroups(ctx context.Context, req *connect.Request[workflow.WatchGroupsRequest], stream *connect.ServerStream[workflow.WatchGroupsResponse]) error {
	logger.Infof(ctx, "Received WatchGroups request")

	groups, err := s.buildTaskGroups(ctx, req.Msg)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}

	// Send initial response with sentinel
	if err := stream.Send(&workflow.WatchGroupsResponse{
		TaskGroups: groups,
		Sentinel:   true,
	}); err != nil {
		return err
	}

	// If EndDate is set, this is a historical query — close the stream
	if req.Msg.EndDate != nil {
		return nil
	}

	// Otherwise poll for updates
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			updated, err := s.buildTaskGroups(ctx, req.Msg)
			if err != nil {
				logger.Warnf(ctx, "Failed to refresh task groups: %v", err)
				continue
			}
			if len(updated) > 0 {
				if err := stream.Send(&workflow.WatchGroupsResponse{
					TaskGroups: updated,
					Sentinel:   true,
				}); err != nil {
					return err
				}
			}
		}
	}
}

// NewRunService creates a new RunService instance
func NewRunService(repo interfaces.Repository, queueClient workflowconnect.QueueServiceClient, storagePrefix string, dataStore *storage.DataStore, stateClient *statek8s.StateClient) *RunService {
	return &RunService{
		repo:          repo,
		queueClient:   queueClient,
		storagePrefix: storagePrefix,
		dataStore:     dataStore,
		stateClient:   stateClient,
	}
}

// Ensure we implement the interface
var _ workflowconnect.RunServiceHandler = (*RunService)(nil)

// CreateRun creates a new run
func (s *RunService) CreateRun(
	ctx context.Context,
	req *connect.Request[workflow.CreateRunRequest],
) (*connect.Response[workflow.CreateRunResponse], error) {
	logger.Infof(ctx, "Received CreateRun request")

	// Validate request
	if err := req.Msg.Validate(); err != nil {
		logger.Errorf(ctx, "Invalid CreateRun request: %v", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Resolve run identity from request (mirrors repo logic for name generation)
	var org, project, domain, name string
	switch id := req.Msg.Id.(type) {
	case *workflow.CreateRunRequest_RunId:
		org = id.RunId.Org
		project = id.RunId.Project
		domain = id.RunId.Domain
		name = id.RunId.Name
	case *workflow.CreateRunRequest_ProjectId:
		org = id.ProjectId.Organization
		project = id.ProjectId.Name
		domain = id.ProjectId.Domain
		name = fmt.Sprintf("run-%d", time.Now().Unix())
	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid run ID type"))
	}

	// Compute storage URIs before DB insert so they're persisted in the ActionSpec
	inputPrefix := buildInputPrefix(s.storagePrefix, org, project, domain, name)
	runOutputBase := buildRunOutputBase(s.storagePrefix, org, project, domain, name)

	// Persist inputs to storage
	if req.Msg.Inputs != nil && len(req.Msg.Inputs.Literals) > 0 {
		literalMap := inputsToLiteralMap(req.Msg.Inputs)
		inputRef := storage.DataReference(inputPrefix + "/inputs.pb")
		if err := s.dataStore.WriteProtobuf(ctx, inputRef, storage.Options{}, literalMap); err != nil {
			logger.Errorf(ctx, "Failed to write inputs to storage: %v", err)
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to write inputs: %w", err))
		}
		logger.Infof(ctx, "Wrote inputs to %s", inputRef)
	}

	// Create run in database with storage URIs
	run, err := s.repo.ActionRepo().CreateRun(ctx, req.Msg, inputPrefix, runOutputBase)
	if err != nil {
		logger.Errorf(ctx, "Failed to create run: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Enqueue the root action to the queue service
	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     run.Org,
			Project: run.Project,
			Domain:  run.Domain,
			Name:    run.Name,
		},
		Name: run.Name, // For root actions, action name = run name
	}

	// Build EnqueueActionRequest from CreateRunRequest
	enqueueReq := &workflow.EnqueueActionRequest{
		ActionId:      actionID,
		RunSpec:       req.Msg.RunSpec,
		InputUri:      inputPrefix,
		RunOutputBase: runOutputBase,
	}

	// Set the spec based on the task type in CreateRunRequest
	switch taskSpec := req.Msg.Task.(type) {
	case *workflow.CreateRunRequest_TaskSpec:
		enqueueReq.Spec = &workflow.EnqueueActionRequest_Task{
			Task: &workflow.TaskAction{
				Spec: taskSpec.TaskSpec,
			},
		}
	case *workflow.CreateRunRequest_TaskId:
		enqueueReq.Spec = &workflow.EnqueueActionRequest_Task{
			Task: &workflow.TaskAction{
				Id: taskSpec.TaskId,
			},
		}
	}

	// Call queue service to enqueue the root action
	_, err = s.queueClient.EnqueueAction(ctx, connect.NewRequest(enqueueReq))
	if err != nil {
		logger.Errorf(ctx, "Failed to enqueue root action: %v", err)
		// Note: We don't fail the CreateRun if enqueue fails - the run is already created
		// In production, you might want to mark the run as failed or retry the enqueue
		logger.Warnf(ctx, "Run %s created but failed to enqueue root action", run.Name)
	} else {
		logger.Infof(ctx, "Successfully enqueued root action for run %s", run.Name)
	}

	// Build response (simplified - you'd convert the full Run model)
	resp := &workflow.CreateRunResponse{
		Run: &workflow.Run{
			Action: &workflow.Action{
				Id: actionID,
			},
		},
	}

	return connect.NewResponse(resp), nil
}

// AbortRun aborts a run
func (s *RunService) AbortRun(
	ctx context.Context,
	req *connect.Request[workflow.AbortRunRequest],
) (*connect.Response[workflow.AbortRunResponse], error) {
	logger.Infof(ctx, "Received AbortRun request for run: %s/%s/%s/%s",
		req.Msg.RunId.Org, req.Msg.RunId.Project, req.Msg.RunId.Domain, req.Msg.RunId.Name)

	// Validate request
	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	reason := "User requested abort"
	if req.Msg.Reason != nil {
		reason = *req.Msg.Reason
	}

	// Abort in database
	if err := s.repo.ActionRepo().AbortRun(ctx, req.Msg.RunId, reason, nil); err != nil {
		logger.Errorf(ctx, "Failed to abort run: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&workflow.AbortRunResponse{}), nil
}

// GetRunDetails gets detailed information about a run, combining DB and K8s data.
func (s *RunService) GetRunDetails(
	ctx context.Context,
	req *connect.Request[workflow.GetRunDetailsRequest],
) (*connect.Response[workflow.GetRunDetailsResponse], error) {
	logger.Infof(ctx, "Received GetRunDetails request")

	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Get run from database
	run, err := s.repo.ActionRepo().GetRun(ctx, req.Msg.RunId)
	if err != nil {
		logger.Errorf(ctx, "Failed to get run: %v", err)
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	details := &workflow.RunDetails{}

	// Try to get live action details from K8s
	rootActionID := &common.ActionIdentifier{
		Run:  req.Msg.RunId,
		Name: run.Name, // For root actions, action name = run name
	}
	ta, err := s.stateClient.GetTaskAction(ctx, rootActionID)
	if err != nil {
		// K8s CR may not exist yet — fall back to DB-only info
		logger.Infof(ctx, "TaskAction not found in K8s, using DB data for run: %s", run.Name)
		details.Action = &workflow.ActionDetails{
			Id: rootActionID,
			Status: &workflow.ActionStatus{
				Phase:     actionPhaseFromString(run.Phase),
				StartTime: timestamppb.New(run.CreatedAt),
			},
		}
	} else {
		actionDetails, err := s.taskActionToActionDetails(ta)
		if err != nil {
			logger.Warnf(ctx, "Failed to convert TaskAction for run details: %v", err)
		} else {
			details.Action = actionDetails
		}
	}

	logger.Infof(ctx, "Retrieved run details for: %s", run.Name)
	return connect.NewResponse(&workflow.GetRunDetailsResponse{Details: details}), nil
}

// GetActionDetails gets detailed information about an action from K8s.
func (s *RunService) GetActionDetails(
	ctx context.Context,
	req *connect.Request[workflow.GetActionDetailsRequest],
) (*connect.Response[workflow.GetActionDetailsResponse], error) {
	logger.Infof(ctx, "Received GetActionDetails request")

	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	ta, err := s.stateClient.GetTaskAction(ctx, req.Msg.ActionId)
	if err != nil {
		logger.Errorf(ctx, "Failed to get TaskAction: %v", err)
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("TaskAction not found: %w", err))
	}

	details, err := s.taskActionToActionDetails(ta)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to convert TaskAction: %w", err))
	}

	logger.Infof(ctx, "Retrieved action details for: %s", req.Msg.ActionId.Name)
	return connect.NewResponse(&workflow.GetActionDetailsResponse{Details: details}), nil
}

// GetActionData gets input and output data for an action by reading from storage.
func (s *RunService) GetActionData(
	ctx context.Context,
	req *connect.Request[workflow.GetActionDataRequest],
) (*connect.Response[workflow.GetActionDataResponse], error) {
	logger.Infof(ctx, "Received GetActionData request for: %s/%s",
		req.Msg.ActionId.Run.Name, req.Msg.ActionId.Name)

	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Get TaskAction CR for storage URIs
	ta, err := s.stateClient.GetTaskAction(ctx, req.Msg.ActionId)
	if err != nil {
		logger.Errorf(ctx, "Failed to get TaskAction: %v", err)
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("TaskAction not found: %w", err))
	}

	resp := &workflow.GetActionDataResponse{
		Inputs:  &task.Inputs{},
		Outputs: &task.Outputs{},
	}

	// Read inputs from storage
	if ta.Spec.InputURI != "" {
		inputRef := storage.DataReference(ta.Spec.InputURI)
		logger.Debugf(ctx, "Reading inputs from: %s", inputRef)
		inputMap := &core.LiteralMap{}
		if err := s.dataStore.ReadProtobuf(ctx, inputRef, inputMap); err != nil {
			if !storage.IsNotFound(err) {
				logger.Errorf(ctx, "Failed to read inputs from %s: %v", inputRef, err)
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to read inputs: %w", err))
			}
			logger.Debugf(ctx, "Inputs not found at %s", inputRef)
		} else {
			resp.Inputs = literalMapToInputs(inputMap)
			logger.Debugf(ctx, "Read %d input literals", len(resp.Inputs.Literals))
		}
	} else {
		logger.Warnf(ctx, "TaskAction %s has empty InputURI", ta.Spec.ActionName)
	}

	// Read outputs from storage (only present if action succeeded)
	outputBase := actionOutputURI(ta)
	if outputBase != "" {
		outputBaseRef := storage.DataReference(outputBase)
		outputRef, err := s.dataStore.ConstructReference(ctx, outputBaseRef, "outputs.pb")
		if err != nil {
			logger.Errorf(ctx, "Failed to construct output reference from %s: %v", outputBase, err)
			// Don't fail — outputs are optional
		} else {
			logger.Debugf(ctx, "Reading outputs from: %s", outputRef)
			outputMap := &core.LiteralMap{}
			if err := s.dataStore.ReadProtobuf(ctx, outputRef, outputMap); err != nil {
				if !storage.IsNotFound(err) {
					logger.Errorf(ctx, "Failed to read outputs from %s: %v", outputRef, err)
				} else {
					logger.Debugf(ctx, "Outputs not found at %s (action may not have finished)", outputRef)
				}
			} else {
				resp.Outputs = literalMapToOutputs(outputMap)
				logger.Debugf(ctx, "Read %d output literals", len(resp.Outputs.Literals))
			}
		}
	} else {
		logger.Warnf(ctx, "TaskAction %s has empty RunOutputBase", ta.Spec.ActionName)
	}

	logger.Infof(ctx, "Retrieved action data for: %s (inputs=%d, outputs=%d)",
		req.Msg.ActionId.Name, len(resp.Inputs.Literals), len(resp.Outputs.Literals))
	return connect.NewResponse(resp), nil
}

// ListRuns lists runs based on filter criteria
func (s *RunService) ListRuns(
	ctx context.Context,
	req *connect.Request[workflow.ListRunsRequest],
) (*connect.Response[workflow.ListRunsResponse], error) {
	logger.Infof(ctx, "Received ListRuns request")

	// Validate request
	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// List runs from database
	runs, nextToken, err := s.repo.ActionRepo().ListRuns(ctx, req.Msg)
	if err != nil {
		logger.Errorf(ctx, "Failed to list runs: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert to proto format
	protoRuns := make([]*workflow.Run, len(runs))
	for i, run := range runs {
		protoRuns[i] = s.convertRunToProto(run)
	}

	resp := &workflow.ListRunsResponse{
		Runs:  protoRuns,
		Token: nextToken,
	}

	logger.Infof(ctx, "Listed %d runs", len(runs))
	return connect.NewResponse(resp), nil
}

// ListActions lists actions for a run
func (s *RunService) ListActions(
	ctx context.Context,
	req *connect.Request[workflow.ListActionsRequest],
) (*connect.Response[workflow.ListActionsResponse], error) {
	logger.Infof(ctx, "Received ListActions request")

	// Validate request
	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// List actions from database
	limit := 50
	if req.Msg.Request != nil && req.Msg.Request.Limit > 0 {
		limit = int(req.Msg.Request.Limit)
	}

	actions, nextToken, err := s.repo.ActionRepo().ListActions(ctx, req.Msg.RunId, limit, "")
	if err != nil {
		logger.Errorf(ctx, "Failed to list actions: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert to proto format
	protoActions := make([]*workflow.Action, len(actions))
	for i, action := range actions {
		ea := s.convertActionToEnrichedProto(action)
		protoActions[i] = ea.Action
	}

	resp := &workflow.ListActionsResponse{
		Actions: protoActions,
		Token:   nextToken,
	}

	logger.Infof(ctx, "Listed %d actions", len(actions))
	return connect.NewResponse(resp), nil
}

// AbortAction aborts a specific action
func (s *RunService) AbortAction(
	ctx context.Context,
	req *connect.Request[workflow.AbortActionRequest],
) (*connect.Response[workflow.AbortActionResponse], error) {
	logger.Infof(ctx, "Received AbortAction request")

	// Validate request
	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	reason := "User requested abort"
	if req.Msg.Reason != "" {
		reason = req.Msg.Reason
	}

	// Abort in database
	if err := s.repo.ActionRepo().AbortAction(ctx, req.Msg.ActionId, reason, nil); err != nil {
		logger.Errorf(ctx, "Failed to abort action: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&workflow.AbortActionResponse{}), nil
}

// Streaming RPCs (simplified implementations)

// WatchRunDetails streams run details updates
func (s *RunService) WatchRunDetails(
	ctx context.Context,
	req *connect.Request[workflow.WatchRunDetailsRequest],
	stream *connect.ServerStream[workflow.WatchRunDetailsResponse],
) error {
	logger.Infof(ctx, "Received WatchRunDetails request")

	// For now, just send initial state and close
	// TODO: Implement actual streaming with polling or database triggers
	run, err := s.repo.ActionRepo().GetRun(ctx, req.Msg.RunId)
	if err != nil {
		return connect.NewError(connect.CodeNotFound, err)
	}

	resp := &workflow.WatchRunDetailsResponse{
		Details: &workflow.RunDetails{
			// Would populate from run model
		},
	}

	if err := stream.Send(resp); err != nil {
		return err
	}

	logger.Infof(ctx, "Sent initial run details for: %s", run.Name)

	// Keep connection open and send updates (simplified)
	updates := make(chan *models.Run)
	errs := make(chan error)

	go s.repo.ActionRepo().WatchRunUpdates(ctx, req.Msg.RunId, updates, errs)

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errs:
			return connect.NewError(connect.CodeInternal, err)
		case run := <-updates:
			resp := &workflow.WatchRunDetailsResponse{
				Details: &workflow.RunDetails{
					// Would populate from run
				},
			}
			if err := stream.Send(resp); err != nil {
				return err
			}
			logger.Infof(ctx, "Sent run update for: %s", run.Name)
		}
	}
}

// WatchActionDetails streams action details updates from K8s TaskAction CRs.
// Phase transition history is read from the CRD's PhaseHistory field, which is
// durably maintained by the executor controller.
func (s *RunService) WatchActionDetails(
	ctx context.Context,
	req *connect.Request[workflow.WatchActionDetailsRequest],
	stream *connect.ServerStream[workflow.WatchActionDetailsResponse],
) error {
	actionID := req.Msg.ActionId
	logger.Infof(ctx, "Received WatchActionDetails request for: %s/%s", actionID.Run.Name, actionID.Name)

	// Step 1: Get initial state from K8s
	ta, err := s.stateClient.GetTaskAction(ctx, actionID)
	if err != nil {
		return connect.NewError(connect.CodeNotFound, fmt.Errorf("TaskAction not found: %w", err))
	}

	details, err := s.taskActionToActionDetails(ta)
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to convert TaskAction: %w", err))
	}

	if err := stream.Send(&workflow.WatchActionDetailsResponse{Details: details}); err != nil {
		return err
	}

	// Step 2: Subscribe and stream updates
	updatesCh := s.stateClient.Subscribe()
	defer s.stateClient.Unsubscribe(updatesCh)

	for {
		select {
		case <-ctx.Done():
			return nil
		case update, ok := <-updatesCh:
			if !ok {
				return nil
			}
			// Filter to this specific action
			aid := update.ActionID
			if aid.Run.Org != actionID.Run.Org || aid.Run.Project != actionID.Run.Project ||
				aid.Run.Domain != actionID.Run.Domain || aid.Run.Name != actionID.Run.Name ||
				aid.Name != actionID.Name {
				continue
			}

			// Re-fetch the full CR for complete PhaseHistory
			ta, err := s.stateClient.GetTaskAction(ctx, actionID)
			if err != nil {
				logger.Warnf(ctx, "Failed to re-fetch TaskAction: %v", err)
				continue
			}

			details, err := s.taskActionToActionDetails(ta)
			if err != nil {
				logger.Warnf(ctx, "Failed to convert TaskAction: %v", err)
				continue
			}

			if err := stream.Send(&workflow.WatchActionDetailsResponse{Details: details}); err != nil {
				return err
			}
		}
	}
}

// WatchRuns streams run updates based on filter criteria
func (s *RunService) WatchRuns(
	ctx context.Context,
	req *connect.Request[workflow.WatchRunsRequest],
	stream *connect.ServerStream[workflow.WatchRunsResponse],
) error {
	logger.Infof(ctx, "Received WatchRuns request")

	// Step 1: Send existing runs that match filter
	listReq := s.convertWatchRequestToListRequest(req.Msg)

	runs, _, err := s.repo.ActionRepo().ListRuns(ctx, listReq)
	if err != nil {
		logger.Errorf(ctx, "Failed to list runs: %v", err)
		// Continue even if list fails - still watch for new updates
	} else if len(runs) > 0 {
		// Send existing runs
		protoRuns := make([]*workflow.Run, len(runs))
		for i, run := range runs {
			protoRuns[i] = s.convertRunToProto(run)
		}

		if err := stream.Send(&workflow.WatchRunsResponse{
			Runs: protoRuns,
		}); err != nil {
			return err
		}
	}

	// Step 2: Watch for run updates using repository notifications
	// Create channels for receiving updates
	updatesCh := make(chan *models.Run, 10)
	errsCh := make(chan error, 1)

	// Start watching for updates in a goroutine
	go s.repo.ActionRepo().WatchAllRunUpdates(ctx, updatesCh, errsCh)

	for {
		select {
		case <-ctx.Done():
			return nil

		case err := <-errsCh:
			logger.Errorf(ctx, "Error watching runs: %v", err)
			return err

		case run := <-updatesCh:
			// Filter the run based on the watch request criteria
			if !s.runMatchesFilter(run, req.Msg) {
				continue
			}

			// Convert and send the updated run
			protoRun := s.convertRunToProto(run)
			if err := stream.Send(&workflow.WatchRunsResponse{
				Runs: []*workflow.Run{protoRun},
			}); err != nil {
				return err
			}
		}
	}
}

// WatchActions streams action updates for a run by watching Kubernetes TaskAction CRs.
func (s *RunService) WatchActions(
	ctx context.Context,
	req *connect.Request[workflow.WatchActionsRequest],
	stream *connect.ServerStream[workflow.WatchActionsResponse],
) error {
	runID := req.Msg.RunId
	logger.Infof(ctx, "Received WatchActions request for run: %s", runID.Name)

	// Step 1: Subscribe to K8s watch events, filter to this run
	updatesCh := s.stateClient.Subscribe()
	defer s.stateClient.Unsubscribe(updatesCh)

	// Step 2: Send existing TaskAction CRs for this run
	taskActions, err := s.stateClient.ListRunActions(ctx, runID)
	if err != nil {
		logger.Errorf(ctx, "Failed to list TaskAction CRs: %v", err)
		// Continue — still watch for new updates
	} else if len(taskActions) > 0 {
		enriched := make([]*workflow.EnrichedAction, 0, len(taskActions))
		for _, ta := range taskActions {
			enriched = append(enriched, s.taskActionToEnrichedProto(ta))
		}
		if err := stream.Send(&workflow.WatchActionsResponse{
			EnrichedActions: enriched,
		}); err != nil {
			return err
		}
	}

	for {
		select {
		case <-ctx.Done():
			return nil

		case update, ok := <-updatesCh:
			if !ok {
				return nil
			}
			// Filter to this run
			aid := update.ActionID
			if aid.Run.Org != runID.Org || aid.Run.Project != runID.Project ||
				aid.Run.Domain != runID.Domain || aid.Run.Name != runID.Name {
				continue
			}

			metadata := &workflow.ActionMetadata{
				ActionType: workflow.ActionType_ACTION_TYPE_TASK,
				Spec: &workflow.ActionMetadata_Task{
					Task: &workflow.TaskActionMetadata{
						TaskType:  update.TaskType,
						ShortName: update.ShortName,
					},
				},
			}
			if update.ParentActionName != "" {
				metadata.Parent = update.ParentActionName
			}

			enriched := &workflow.EnrichedAction{
				Action: &workflow.Action{
					Id:       aid,
					Metadata: metadata,
					Status: &workflow.ActionStatus{
						Phase: actionPhaseFromString(update.Phase),
					},
				},
				MeetsFilter: !update.IsDeleted,
			}
			if err := stream.Send(&workflow.WatchActionsResponse{
				EnrichedActions: []*workflow.EnrichedAction{enriched},
			}); err != nil {
				return err
			}
		}
	}
}

// WatchClusterEvents streams cluster events derived from TaskAction condition transitions.
func (s *RunService) WatchClusterEvents(
	ctx context.Context,
	req *connect.Request[workflow.WatchClusterEventsRequest],
	stream *connect.ServerStream[workflow.WatchClusterEventsResponse],
) error {
	actionID := req.Msg.Id
	logger.Infof(ctx, "Received WatchClusterEvents request for: %s/%s", actionID.Run.Name, actionID.Name)

	// Step 1: Get initial conditions and send existing events
	ta, err := s.stateClient.GetTaskAction(ctx, actionID)
	if err != nil {
		return connect.NewError(connect.CodeNotFound, fmt.Errorf("TaskAction not found: %w", err))
	}

	lastConditions := ta.Status.Conditions
	events := conditionsToClusterEvents(lastConditions)
	if len(events) > 0 {
		if err := stream.Send(&workflow.WatchClusterEventsResponse{ClusterEvents: events}); err != nil {
			return err
		}
	}

	// Step 2: Subscribe and stream new events as conditions change
	updatesCh := s.stateClient.Subscribe()
	defer s.stateClient.Unsubscribe(updatesCh)

	for {
		select {
		case <-ctx.Done():
			return nil
		case update, ok := <-updatesCh:
			if !ok {
				return nil
			}
			// Filter to this specific action
			aid := update.ActionID
			if aid.Run.Org != actionID.Run.Org || aid.Run.Project != actionID.Run.Project ||
				aid.Run.Domain != actionID.Run.Domain || aid.Run.Name != actionID.Run.Name ||
				aid.Name != actionID.Name {
				continue
			}

			// Re-fetch for full conditions
			ta, err := s.stateClient.GetTaskAction(ctx, actionID)
			if err != nil {
				logger.Warnf(ctx, "Failed to re-fetch TaskAction for events: %v", err)
				continue
			}

			// Find new events by comparing condition counts
			newEvents := diffClusterEvents(lastConditions, ta.Status.Conditions)
			if len(newEvents) == 0 {
				continue
			}
			lastConditions = ta.Status.Conditions

			if err := stream.Send(&workflow.WatchClusterEventsResponse{ClusterEvents: newEvents}); err != nil {
				return err
			}
		}
	}
}

// diffClusterEvents returns ClusterEvents for conditions that are new (True) in current but not in prev.
func diffClusterEvents(prev, current []metav1.Condition) []*workflow.ClusterEvent {
	prevSet := make(map[string]bool, len(prev))
	for _, c := range prev {
		if c.Status == metav1.ConditionTrue {
			prevSet[c.Type+"/"+c.Reason] = true
		}
	}

	var events []*workflow.ClusterEvent
	for _, c := range current {
		if c.Status != metav1.ConditionTrue {
			continue
		}
		key := c.Type + "/" + c.Reason
		if prevSet[key] {
			continue
		}
		msg := c.Type
		if c.Reason != "" {
			msg += ": " + c.Reason
		}
		if c.Message != "" {
			msg += " - " + c.Message
		}
		events = append(events, &workflow.ClusterEvent{
			OccurredAt: timestamppb.New(c.LastTransitionTime.Time),
			Message:    msg,
		})
	}
	return events
}

// Helper functions

// buildInputPrefix generates the input path prefix for the root action.
// The executor appends "inputs.pb" to this prefix when reading.
// Example: s3://bucket/org/project/domain/run-name/inputs
func buildInputPrefix(storagePrefix, org, project, domain, name string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s/inputs",
		strings.TrimRight(storagePrefix, "/"),
		org, project, domain, name)
}

// inputsToLiteralMap converts task.Inputs (ordered NamedLiteral list) to core.LiteralMap (map).
func inputsToLiteralMap(inputs *task.Inputs) *core.LiteralMap {
	m := make(map[string]*core.Literal, len(inputs.Literals))
	for _, nl := range inputs.Literals {
		m[nl.Name] = nl.Value
	}
	return &core.LiteralMap{Literals: m}
}

// literalMapToInputs converts a core.LiteralMap to task.Inputs (reverse of inputsToLiteralMap).
func literalMapToInputs(m *core.LiteralMap) *task.Inputs {
	if m == nil || len(m.Literals) == 0 {
		return &task.Inputs{}
	}
	literals := make([]*task.NamedLiteral, 0, len(m.Literals))
	for name, val := range m.Literals {
		literals = append(literals, &task.NamedLiteral{Name: name, Value: val})
	}
	sort.Slice(literals, func(i, j int) bool { return literals[i].Name < literals[j].Name })
	return &task.Inputs{Literals: literals}
}

// literalMapToOutputs converts a core.LiteralMap to task.Outputs.
func literalMapToOutputs(m *core.LiteralMap) *task.Outputs {
	if m == nil || len(m.Literals) == 0 {
		return &task.Outputs{}
	}
	literals := make([]*task.NamedLiteral, 0, len(m.Literals))
	for name, val := range m.Literals {
		literals = append(literals, &task.NamedLiteral{Name: name, Value: val})
	}
	sort.Slice(literals, func(i, j int) bool { return literals[i].Name < literals[j].Name })
	return &task.Outputs{Literals: literals}
}

// actionOutputURI computes the action-specific output URI from the TaskAction spec.
func actionOutputURI(ta *executorv1.TaskAction) string {
	if ta.Spec.RunOutputBase == "" {
		return ""
	}
	return strings.TrimRight(ta.Spec.RunOutputBase, "/") + "/" + ta.Spec.ActionName
}

// buildRunOutputBase generates the output base path for the run.
// Example: s3://bucket/org/project/domain/run-name/
func buildRunOutputBase(storagePrefix, org, project, domain, name string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s/",
		strings.TrimRight(storagePrefix, "/"),
		org, project, domain, name)
}

// convertRunToProto converts a repository Run to a proto Run
func (s *RunService) convertRunToProto(run *models.Run) *workflow.Run {
	if run == nil {
		return nil
	}

	// Build the action identifier from the run
	runID := &common.RunIdentifier{
		Org:     run.Org,
		Project: run.Project,
		Domain:  run.Domain,
		Name:    run.Name,
	}

	// Create the root action with status
	action := &workflow.Action{
		Id: &common.ActionIdentifier{
			Run:  runID,
			Name: run.Name, // For root actions, action name = run name
		},
		Metadata: &workflow.ActionMetadata{
			// TODO: Extract from ActionSpec JSON if needed
		},
		Status: &workflow.ActionStatus{
			Phase: common.ActionPhase(common.ActionPhase_value[run.Phase]),
			// TODO: Extract timestamps, error, etc. from ActionDetails JSON
		},
	}

	return &workflow.Run{
		Action: action,
	}
}

// convertActionToEnrichedProto converts a repository Action to an EnrichedAction proto
func (s *RunService) convertActionToEnrichedProto(action *models.Action) *workflow.EnrichedAction {
	if action == nil {
		return nil
	}

	// Build the action identifier
	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     action.Org,
			Project: action.Project,
			Domain:  action.Domain,
			Name:    action.GetRunName(),
		},
		Name: action.Name,
	}

	// Build the action status
	actionStatus := &workflow.ActionStatus{
		Phase: common.ActionPhase(common.ActionPhase_value[action.Phase]),
	}

	// Build the action proto
	actionProto := &workflow.Action{
		Id:     actionID,
		Status: actionStatus,
	}

	return &workflow.EnrichedAction{
		Action:      actionProto,
		MeetsFilter: true, // For now, all actions meet the filter
	}
}

// taskActionToEnrichedProto converts a Kubernetes TaskAction CR to an EnrichedAction proto.
func (s *RunService) taskActionToEnrichedProto(ta *executorv1.TaskAction) *workflow.EnrichedAction {
	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     ta.Spec.Org,
			Project: ta.Spec.Project,
			Domain:  ta.Spec.Domain,
			Name:    ta.Spec.RunName,
		},
		Name: ta.Spec.ActionName,
	}

	// Determine short name: use spec.ShortName if set, otherwise extract from template ID
	shortName := ta.Spec.ShortName
	if shortName == "" && len(ta.Spec.TaskTemplate) > 0 {
		tmpl := &core.TaskTemplate{}
		if err := proto.Unmarshal(ta.Spec.TaskTemplate, tmpl); err == nil && tmpl.GetId() != nil {
			shortName = extractShortName(tmpl.GetId().GetName())
		}
	}

	metadata := &workflow.ActionMetadata{
		ActionType: workflow.ActionType_ACTION_TYPE_TASK,
		Spec: &workflow.ActionMetadata_Task{
			Task: &workflow.TaskActionMetadata{
				TaskType:  ta.Spec.TaskType,
				ShortName: shortName,
			},
		},
	}
	if ta.Spec.ParentActionName != nil {
		metadata.Parent = *ta.Spec.ParentActionName
	}

	phase := actionPhaseFromString(statek8s.GetPhaseFromConditions(ta))
	status := &workflow.ActionStatus{
		Phase:     phase,
		StartTime: timestamppb.New(ta.CreationTimestamp.Time),
	}

	// Derive end time from the terminal condition's LastTransitionTime
	for _, cond := range ta.Status.Conditions {
		if (cond.Type == string(executorv1.ConditionTypeSucceeded) || cond.Type == string(executorv1.ConditionTypeFailed)) &&
			cond.Status == "True" {
			status.EndTime = timestamppb.New(cond.LastTransitionTime.Time)
			break
		}
	}

	return &workflow.EnrichedAction{
		Action: &workflow.Action{
			Id:       actionID,
			Metadata: metadata,
			Status:   status,
		},
		MeetsFilter: true,
	}
}

// actionPhaseFromString maps state-client phase strings (e.g. "PHASE_QUEUED")
// to the proto ActionPhase enum.
func actionPhaseFromString(phase string) common.ActionPhase {
	// The state client returns "PHASE_*" while the proto enum uses "ACTION_PHASE_*".
	mapped := "ACTION_" + phase
	if v, ok := common.ActionPhase_value[mapped]; ok {
		return common.ActionPhase(v)
	}
	return common.ActionPhase_ACTION_PHASE_UNSPECIFIED
}

// buildTaskGroups queries the DB for root actions and groups them by task name in memory.
func (s *RunService) buildTaskGroups(ctx context.Context, req *workflow.WatchGroupsRequest) ([]*workflow.TaskGroup, error) {
	var org, project, domain string
	if pid := req.GetProjectId(); pid != nil {
		org = pid.Organization
		project = pid.Name
		domain = pid.Domain
	}

	var startDate, endDate *time.Time
	if req.StartDate != nil {
		t := req.StartDate.AsTime()
		startDate = &t
	}
	if req.EndDate != nil {
		t := req.EndDate.AsTime()
		endDate = &t
	}

	actions, err := s.repo.ActionRepo().ListRootActions(ctx, org, project, domain, startDate, endDate, 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to list root actions: %w", err)
	}

	// Group by task name extracted from ActionSpec JSON
	type recentAction struct {
		runName string
		phase   common.ActionPhase
		created time.Time
	}
	type groupAccum struct {
		taskName      string
		envName       string
		shortName     string
		count         int64
		latestTime    time.Time
		phaseCounts   map[common.ActionPhase]int64
		failCount     int64
		recentActions []recentAction
	}
	groups := make(map[string]*groupAccum)

	for _, action := range actions {
		taskName := extractTaskName(action.ActionSpec)
		if taskName == "" {
			taskName = action.Name // fallback to action name
		}

		g, ok := groups[taskName]
		if !ok {
			g = &groupAccum{
				taskName:    taskName,
				phaseCounts: make(map[common.ActionPhase]int64),
			}
			groups[taskName] = g
		}
		g.count++
		if action.CreatedAt.After(g.latestTime) {
			g.latestTime = action.CreatedAt
		}

		phase := common.ActionPhase(common.ActionPhase_value["ACTION_"+action.Phase])
		g.phaseCounts[phase]++
		if phase == common.ActionPhase_ACTION_PHASE_FAILED {
			g.failCount++
		}

		// Track recent actions (keep up to 10 most recent per group)
		g.recentActions = append(g.recentActions, recentAction{
			runName: action.GetRunName(),
			phase:   phase,
			created: action.CreatedAt,
		})
	}

	// Convert to proto
	result := make([]*workflow.TaskGroup, 0, len(groups))
	for _, g := range groups {
		var phaseCounts []*workflow.TaskGroup_PhaseCounts
		for phase, count := range g.phaseCounts {
			phaseCounts = append(phaseCounts, &workflow.TaskGroup_PhaseCounts{
				Phase: phase,
				Count: count,
			})
		}

		var failRate float64
		if g.count > 0 {
			failRate = float64(g.failCount) / float64(g.count)
		}

		// Build RecentStatuses sorted newest-first, capped at 10
		sort.Slice(g.recentActions, func(i, j int) bool {
			return g.recentActions[i].created.After(g.recentActions[j].created)
		})
		limit := 10
		if len(g.recentActions) < limit {
			limit = len(g.recentActions)
		}
		recentStatuses := make([]*workflow.TaskGroup_RecentStatus, limit)
		for i := 0; i < limit; i++ {
			recentStatuses[i] = &workflow.TaskGroup_RecentStatus{
				RunName: g.recentActions[i].runName,
				Phase:   g.recentActions[i].phase,
			}
		}

		tg := &workflow.TaskGroup{
			TaskName:           g.taskName,
			EnvironmentName:    g.envName,
			ShortName:          g.shortName,
			TotalRuns:          g.count,
			LatestRunTime:      timestamppb.New(g.latestTime),
			AverageFailureRate: failRate,
			PhaseCounts:        phaseCounts,
			RecentStatuses:     recentStatuses,
		}
		result = append(result, tg)
	}
	return result, nil
}

// extractTaskName attempts to extract the task name from an ActionSpec JSON blob.
func extractTaskName(specJSON []byte) string {
	if len(specJSON) == 0 {
		return ""
	}
	// Use a lightweight struct to avoid full proto deserialization
	var spec struct {
		Spec *struct {
			Task *struct {
				ID *struct {
					Name string `json:"name"`
				} `json:"id"`
				Spec *struct {
					ShortName string `json:"short_name"`
				} `json:"spec"`
			} `json:"task"`
		} `json:"spec"`
	}
	if err := json.Unmarshal(specJSON, &spec); err != nil {
		return ""
	}
	if spec.Spec != nil && spec.Spec.Task != nil {
		if spec.Spec.Task.ID != nil && spec.Spec.Task.ID.Name != "" {
			return spec.Spec.Task.ID.Name
		}
		if spec.Spec.Task.Spec != nil && spec.Spec.Task.Spec.ShortName != "" {
			return spec.Spec.Task.Spec.ShortName
		}
	}
	return ""
}

// extractShortName extracts a human-readable function name from a task template ID name.
// It removes any environment name prefix (format: "envName.functionName") or if no prefix,
// splits on '.' and returns the last part.
func extractShortName(name string) string {
	if name == "" {
		return ""
	}
	// Split on '.' and take the last part
	parts := strings.Split(name, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return name
}

// taskActionToActionDetails converts a K8s TaskAction CR to a full ActionDetails proto.
func (s *RunService) taskActionToActionDetails(ta *executorv1.TaskAction) (*workflow.ActionDetails, error) {
	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     ta.Spec.Org,
			Project: ta.Spec.Project,
			Domain:  ta.Spec.Domain,
			Name:    ta.Spec.RunName,
		},
		Name: ta.Spec.ActionName,
	}

	// Deserialize TaskTemplate from spec first (needed for shortName fallback)
	var tmpl *core.TaskTemplate
	if len(ta.Spec.TaskTemplate) > 0 {
		tmpl = &core.TaskTemplate{}
		if err := proto.Unmarshal(ta.Spec.TaskTemplate, tmpl); err != nil {
			logger.Warnf(context.Background(), "Failed to unmarshal TaskTemplate: %v", err)
			tmpl = nil
		}
	}

	// Determine short name: use spec.ShortName if set, otherwise extract from template ID
	shortName := ta.Spec.ShortName
	if shortName == "" && tmpl != nil && tmpl.GetId() != nil {
		shortName = extractShortName(tmpl.GetId().GetName())
	}

	// Build metadata
	metadata := &workflow.ActionMetadata{
		ActionType: workflow.ActionType_ACTION_TYPE_TASK,
		Spec: &workflow.ActionMetadata_Task{
			Task: &workflow.TaskActionMetadata{
				TaskType:  ta.Spec.TaskType,
				ShortName: shortName,
			},
		},
	}
	if ta.Spec.ParentActionName != nil {
		metadata.Parent = *ta.Spec.ParentActionName
	}

	// Build status from conditions
	phase := actionPhaseFromString(statek8s.GetPhaseFromConditions(ta))
	status := &workflow.ActionStatus{
		Phase:     phase,
		StartTime: timestamppb.New(ta.CreationTimestamp.Time),
		Attempts:  1,
	}
	for _, cond := range ta.Status.Conditions {
		if (cond.Type == string(executorv1.ConditionTypeSucceeded) || cond.Type == string(executorv1.ConditionTypeFailed)) &&
			cond.Status == metav1.ConditionTrue {
			status.EndTime = timestamppb.New(cond.LastTransitionTime.Time)
			break
		}
	}

	// Build details
	details := &workflow.ActionDetails{
		Id:       actionID,
		Metadata: metadata,
		Status:   status,
	}

	if tmpl != nil {
		details.Spec = &workflow.ActionDetails_Task{
			Task: &task.TaskSpec{TaskTemplate: tmpl},
		}
	}

	// Build attempt
	attempt := &workflow.ActionAttempt{
		Attempt:   1,
		Phase:     phase,
		StartTime: timestamppb.New(ta.CreationTimestamp.Time),
	}
	if status.EndTime != nil {
		attempt.EndTime = status.EndTime
	}

	// Build phase transitions from the CRD's durable PhaseHistory.
	// If PhaseHistory is empty (pre-existing CRDs), fall back to a synthetic
	// transition from the current phase so the UI always has at least one entry.
	attempt.PhaseTransitions = phaseHistoryToTransitions(ta.Status.PhaseHistory)
	if len(attempt.PhaseTransitions) == 0 {
		attempt.PhaseTransitions = []*workflow.PhaseTransition{{
			Phase:     phase,
			StartTime: timestamppb.New(ta.CreationTimestamp.Time),
			EndTime:   status.EndTime,
		}}
	}

	// Build cluster events from PhaseHistory (each transition is also an event)
	attempt.ClusterEvents = phaseHistoryToClusterEvents(ta.Status.PhaseHistory)

	details.Attempts = []*workflow.ActionAttempt{attempt}

	return details, nil
}

// phaseHistoryToTransitions converts the CRD's durable PhaseHistory to proto PhaseTransitions.
func phaseHistoryToTransitions(history []executorv1.PhaseTransition) []*workflow.PhaseTransition {
	transitions := make([]*workflow.PhaseTransition, 0, len(history))
	for i, h := range history {
		phase := phaseReasonToActionPhase(h.Phase)
		t := &workflow.PhaseTransition{
			Phase:     phase,
			StartTime: timestamppb.New(h.OccurredAt.Time),
		}
		// Set EndTime to the start of the next transition
		if i+1 < len(history) {
			t.EndTime = timestamppb.New(history[i+1].OccurredAt.Time)
		}
		transitions = append(transitions, t)
	}
	return transitions
}

// phaseHistoryToClusterEvents converts PhaseHistory entries to ClusterEvent messages.
func phaseHistoryToClusterEvents(history []executorv1.PhaseTransition) []*workflow.ClusterEvent {
	events := make([]*workflow.ClusterEvent, 0, len(history))
	for _, h := range history {
		msg := h.Phase
		if h.Message != "" {
			msg += ": " + h.Message
		}
		events = append(events, &workflow.ClusterEvent{
			OccurredAt: timestamppb.New(h.OccurredAt.Time),
			Message:    msg,
		})
	}
	return events
}

// phaseReasonToActionPhase maps a condition reason string to a proto ActionPhase.
func phaseReasonToActionPhase(reason string) common.ActionPhase {
	switch reason {
	case string(executorv1.ConditionReasonQueued):
		return common.ActionPhase_ACTION_PHASE_QUEUED
	case string(executorv1.ConditionReasonInitializing):
		return common.ActionPhase_ACTION_PHASE_INITIALIZING
	case string(executorv1.ConditionReasonExecuting):
		return common.ActionPhase_ACTION_PHASE_RUNNING
	case string(executorv1.ConditionReasonCompleted):
		return common.ActionPhase_ACTION_PHASE_SUCCEEDED
	case string(executorv1.ConditionReasonPermanentFailure), string(executorv1.ConditionReasonAborted):
		return common.ActionPhase_ACTION_PHASE_FAILED
	case string(executorv1.ConditionReasonRetryableFailure):
		return common.ActionPhase_ACTION_PHASE_FAILED
	default:
		return common.ActionPhase_ACTION_PHASE_UNSPECIFIED
	}
}

// conditionsToClusterEvents converts TaskAction conditions to ClusterEvent messages.
func conditionsToClusterEvents(conditions []metav1.Condition) []*workflow.ClusterEvent {
	var events []*workflow.ClusterEvent
	for _, cond := range conditions {
		if cond.Status != metav1.ConditionTrue {
			continue
		}
		msg := cond.Type
		if cond.Reason != "" {
			msg += ": " + cond.Reason
		}
		if cond.Message != "" {
			msg += " - " + cond.Message
		}
		events = append(events, &workflow.ClusterEvent{
			OccurredAt: timestamppb.New(cond.LastTransitionTime.Time),
			Message:    msg,
		})
	}
	return events
}

// convertWatchRequestToListRequest converts a WatchRunsRequest to a ListRunsRequest
func (s *RunService) convertWatchRequestToListRequest(req *workflow.WatchRunsRequest) *workflow.ListRunsRequest {
	listReq := &workflow.ListRunsRequest{
		Request: &common.ListRequest{
			Limit: 100,
		},
	}

	// Convert the target filter to the appropriate ListRuns scope
	switch target := req.Target.(type) {
	case *workflow.WatchRunsRequest_Org:
		listReq.ScopeBy = &workflow.ListRunsRequest_Org{
			Org: target.Org,
		}
	case *workflow.WatchRunsRequest_ClusterId:
		// Cluster filtering not directly supported in ListRuns, will filter client-side
		// Could be added to ListRuns if needed
	case *workflow.WatchRunsRequest_ProjectId:
		listReq.ScopeBy = &workflow.ListRunsRequest_ProjectId{
			ProjectId: target.ProjectId,
		}
	case *workflow.WatchRunsRequest_TaskId:
		// Task filtering not directly supported in ListRuns, will filter client-side
		// Could be added to ListRuns if needed
	}

	return listReq
}

// runMatchesFilter checks if a run matches the WatchRunsRequest filter criteria
func (s *RunService) runMatchesFilter(run *models.Run, req *workflow.WatchRunsRequest) bool {
	if req.Target == nil {
		// No filter, all runs match
		return true
	}

	switch target := req.Target.(type) {
	case *workflow.WatchRunsRequest_Org:
		return run.Org == target.Org

	case *workflow.WatchRunsRequest_ClusterId:
		// TODO: Add cluster field to Run model if needed
		// For now, accept all runs
		return true

	case *workflow.WatchRunsRequest_ProjectId:
		return run.Org == target.ProjectId.Organization &&
			run.Project == target.ProjectId.Name &&
			run.Domain == target.ProjectId.Domain

	case *workflow.WatchRunsRequest_TaskId:
		// TODO: Need to check if the run was triggered by this task
		// This would require storing task_id in the Run model or querying actions
		// For now, accept all runs
		return true

	default:
		// Unknown filter, accept all runs
		return true
	}
}
