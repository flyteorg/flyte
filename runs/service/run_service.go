package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions/actionsconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// RunService implements the RunServiceHandler interface
type RunService struct {
	repo          interfaces.Repository
	actionsClient actionsconnect.ActionsServiceClient
	storagePrefix string
	dataStore     *storage.DataStore
}

const (
	runIDLength     = 20
	runStringFormat = "r%s"
)

func generateRunName(seed int64) string {
	rand.Seed(seed)
	return fmt.Sprintf(runStringFormat, rand.String(runIDLength-1))
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
func NewRunService(repo interfaces.Repository, actionsClient actionsconnect.ActionsServiceClient, storagePrefix string, dataStore *storage.DataStore) *RunService {
	return &RunService{
		repo:          repo,
		actionsClient: actionsClient,
		storagePrefix: storagePrefix,
		dataStore:     dataStore,
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

	// Resolve run identity from request. When only a ProjectId is given, generate
	// the run name here and normalise the request to a RunId so that the repo
	// receives a single, pre-formed identifier — avoiding a second independent
	// name generation inside the repo layer.
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
		name = generateRunName(time.Now().UnixNano())
		req.Msg.Id = &workflow.CreateRunRequest_RunId{
			RunId: &common.RunIdentifier{
				Org:     org,
				Project: project,
				Domain:  domain,
				Name:    name,
			},
		}
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

	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     run.Org,
			Project: run.Project,
			Domain:  run.Domain,
			Name:    run.Name,
		},
		Name: run.Name, // For root actions, action name = run name
	}

	// Build the root action for ActionsService.Enqueue
	rootAction := &actions.Action{
		ActionId:      actionID,
		InputUri:      inputPrefix,
		RunOutputBase: runOutputBase,
	}
	switch taskSpec := req.Msg.Task.(type) {
	case *workflow.CreateRunRequest_TaskSpec:
		rootAction.Spec = &actions.Action_Task{
			Task: &workflow.TaskAction{Spec: taskSpec.TaskSpec},
		}
	case *workflow.CreateRunRequest_TaskId:
		rootAction.Spec = &actions.Action_Task{
			Task: &workflow.TaskAction{Id: taskSpec.TaskId},
		}
	}

	_, err = s.actionsClient.Enqueue(ctx, connect.NewRequest(&actions.EnqueueRequest{
		Action:  rootAction,
		RunSpec: req.Msg.RunSpec,
	}))
	if err != nil {
		logger.Errorf(ctx, "Failed to enqueue root action: %v", err)
	} else {
		logger.Infof(ctx, "Successfully enqueued root action for run %s", run.Name)
	}

	// TODO(nary): We should set the phase here to ActionPhase_ACTION_PHASE_QUEUED. As the root action phase
	// will be returned to the client for display if they use .wait(), returning ActionPhase_ACTION_PHASE_UNSPECIFIED
	// will cause error.
	// We should do this after we persist the state to the DB, and when run service fully relies on DB and do not get
	// status from the state client directly.

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

	// Abort via actions service
	rootActionID := &common.ActionIdentifier{
		Run:  req.Msg.RunId,
		Name: "a0",
	}
	if _, err := s.actionsClient.Abort(ctx, connect.NewRequest(&actions.AbortRequest{
		ActionId: rootActionID,
		Reason:   &reason,
	})); err != nil {
		logger.Errorf(ctx, "Failed to abort run via actions service: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&workflow.AbortRunResponse{}), nil
}

// GetRunDetails gets detailed information about a run from the DB.
func (s *RunService) GetRunDetails(
	ctx context.Context,
	req *connect.Request[workflow.GetRunDetailsRequest],
) (*connect.Response[workflow.GetRunDetailsResponse], error) {
	logger.Infof(ctx, "Received GetRunDetails request")

	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	run, err := s.repo.ActionRepo().GetRun(ctx, req.Msg.RunId)
	if err != nil {
		logger.Errorf(ctx, "Failed to get run: %v", err)
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	rootActionID := &common.ActionIdentifier{
		Run:  req.Msg.RunId,
		Name: run.Name,
	}
	details := &workflow.RunDetails{
		Action: &workflow.ActionDetails{
			Id: rootActionID,
			Status: &workflow.ActionStatus{
				Phase:     common.ActionPhase(run.Phase),
				StartTime: timestamppb.New(run.CreatedAt),
			},
		},
	}

	logger.Infof(ctx, "Retrieved run details for: %s", run.Name)
	return connect.NewResponse(&workflow.GetRunDetailsResponse{Details: details}), nil
}

// GetActionDetails gets detailed information about an action from the DB.
func (s *RunService) GetActionDetails(
	ctx context.Context,
	req *connect.Request[workflow.GetActionDetailsRequest],
) (*connect.Response[workflow.GetActionDetailsResponse], error) {
	logger.Infof(ctx, "Received GetActionDetails request")

	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	action, err := s.repo.ActionRepo().GetAction(ctx, req.Msg.ActionId)
	if err != nil {
		logger.Errorf(ctx, "Failed to get action: %v", err)
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("action not found: %w", err))
	}

	details := s.actionModelToDetails(action, req.Msg.ActionId)

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

	// Get action from DB for storage URIs
	action, err := s.repo.ActionRepo().GetAction(ctx, req.Msg.ActionId)
	if err != nil {
		logger.Errorf(ctx, "Failed to get action: %v", err)
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("action not found: %w", err))
	}

	inputURI, runOutputBase := extractStorageURIs(action.ActionSpec)

	resp := &workflow.GetActionDataResponse{
		Inputs:  &task.Inputs{},
		Outputs: &task.Outputs{},
	}

	// Read inputs from storage
	if inputURI != "" {
		inputRef := storage.DataReference(inputURI)
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
		logger.Warnf(ctx, "Action %s has empty InputURI", req.Msg.ActionId.Name)
	}

	// Read outputs from storage (only present if action succeeded)
	if runOutputBase != "" {
		// TODO(nary): maybe a better way to parse it
		outputBase := strings.TrimRight(runOutputBase, "/") + "/" + req.Msg.ActionId.Name
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
		logger.Warnf(ctx, "Action %s has empty RunOutputBase", req.Msg.ActionId.Name)
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
		// Todo:
		// safely handle error from json unmarshal should be handled properly.
		// Add a unit test for convertRunToProto function
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

	// Abort via actions service
	if _, err := s.actionsClient.Abort(ctx, connect.NewRequest(&actions.AbortRequest{
		ActionId: req.Msg.ActionId,
		Reason:   &reason,
	})); err != nil {
		logger.Errorf(ctx, "Failed to abort action via actions service: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&workflow.AbortActionResponse{}), nil
}

// WatchRunDetails streams run details updates from the DB.
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

// WatchActionDetails streams action details updates from the DB.
func (s *RunService) WatchActionDetails(
	ctx context.Context,
	req *connect.Request[workflow.WatchActionDetailsRequest],
	stream *connect.ServerStream[workflow.WatchActionDetailsResponse],
) error {
	actionID := req.Msg.ActionId
	logger.Infof(ctx, "Received WatchActionDetails request for: %s/%s", actionID.Run.Name, actionID.Name)

	// Step 1: Send initial state from DB
	action, err := s.repo.ActionRepo().GetAction(ctx, actionID)
	if err != nil {
		return connect.NewError(connect.CodeNotFound, fmt.Errorf("action not found: %w", err))
	}

	if err := stream.Send(&workflow.WatchActionDetailsResponse{
		Details: s.actionModelToDetails(action, actionID),
	}); err != nil {
		return err
	}

	// Step 2: Watch DB for updates
	updates := make(chan *models.Action)
	errs := make(chan error)
	go s.repo.ActionRepo().WatchActionUpdates(ctx, actionID.Run, updates, errs)

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errs:
			return connect.NewError(connect.CodeInternal, err)
		case updated, ok := <-updates:
			if !ok {
				return nil
			}
			// Filter to this specific action
			if updated.Name != actionID.Name {
				continue
			}
			if err := stream.Send(&workflow.WatchActionDetailsResponse{
				Details: s.actionModelToDetails(updated, actionID),
			}); err != nil {
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

	// Step 2: Watch for run updates from DB
	updatesCh := make(chan *models.Run)
	errsCh := make(chan error)
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

// WatchActions streams action updates for a run from the DB.
func (s *RunService) WatchActions(
	ctx context.Context,
	req *connect.Request[workflow.WatchActionsRequest],
	stream *connect.ServerStream[workflow.WatchActionsResponse],
) error {
	runID := req.Msg.RunId
	logger.Infof(ctx, "Received WatchActions request for run: %s", runID.Name)

	// Start watching for updates from DB first to prevent event miss
	updatesCh := make(chan *models.Action)
	errsCh := make(chan error)
	go s.repo.ActionRepo().WatchActionUpdates(ctx, runID, updatesCh, errsCh)

	rsm, err := newRunStateManager(req.Msg.GetFilter())
	if err != nil {
		return err
	}

	if err := s.listAndSendAllActions(ctx, runID, rsm, stream); err != nil {
		logger.Errorf(ctx, "Failed to list actions: %v", err)
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errsCh:
			return err
		case updated, ok := <-updatesCh:
			if !ok {
				return nil
			}
			updates, err := rsm.upsertActions(ctx, []*models.Action{updated})
			if err != nil {
				return err
			}
			if err := s.sendChangedActions(runID, updates, stream); err != nil {
				return err
			}
		}
	}
}

func (s *RunService) listAndSendAllActions(
	ctx context.Context,
	runID *common.RunIdentifier,
	rsm *runStateManager,
	stream *connect.ServerStream[workflow.WatchActionsResponse],
) error {
	token := ""
	for {
		batch, nextToken, err := s.repo.ActionRepo().ListActions(ctx, runID, 100, token)
		if err != nil {
			return err
		}

		updates, err := rsm.upsertActions(ctx, batch)
		if err != nil {
			return err
		}

		if err := s.sendChangedActions(runID, updates, stream); err != nil {
			return err
		}

		if nextToken == "" {
			return nil
		}
		token = nextToken
	}
}

func (s *RunService) sendChangedActions(
	runID *common.RunIdentifier,
	updates []*nodeUpdate,
	stream *connect.ServerStream[workflow.WatchActionsResponse],
) error {
	if len(updates) == 0 {
		return nil
	}

	enriched := make([]*workflow.EnrichedAction, 0, len(updates))
	for _, update := range updates {
		enrichedAction := s.convertNodeUpdateToEnrichedProto(runID, update)
		if enrichedAction == nil {
			continue
		}
		enriched = append(enriched, enrichedAction)
	}

	return stream.Send(&workflow.WatchActionsResponse{
		EnrichedActions: enriched,
	})
}

// WatchClusterEvents streams cluster events from the DB action updates.
func (s *RunService) WatchClusterEvents(
	ctx context.Context,
	req *connect.Request[workflow.WatchClusterEventsRequest],
	stream *connect.ServerStream[workflow.WatchClusterEventsResponse],
) error {
	actionID := req.Msg.Id
	logger.Infof(ctx, "Received WatchClusterEvents request for: %s/%s", actionID.Run.Name, actionID.Name)

	// Step 1: Send existing events from current DB state
	action, err := s.repo.ActionRepo().GetAction(ctx, actionID)
	if err != nil {
		return connect.NewError(connect.CodeNotFound, fmt.Errorf("action not found: %w", err))
	}

	events := actionModelToClusterEvents(action)
	if len(events) > 0 {
		if err := stream.Send(&workflow.WatchClusterEventsResponse{ClusterEvents: events}); err != nil {
			return err
		}
	}

	// Step 2: Watch for updates from DB
	updatesCh := make(chan *models.Action)
	errsCh := make(chan error)
	go s.repo.ActionRepo().WatchActionUpdates(ctx, actionID.Run, updatesCh, errsCh)

	lastPhase := action.Phase

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errsCh:
			return connect.NewError(connect.CodeInternal, err)
		case updated, ok := <-updatesCh:
			if !ok {
				return nil
			}
			// Filter to this specific action
			if updated.Name != actionID.Name {
				continue
			}
			// Only emit an event when phase changes
			if updated.Phase == lastPhase {
				continue
			}
			lastPhase = updated.Phase

			newEvents := actionModelToClusterEvents(updated)
			if len(newEvents) == 0 {
				continue
			}
			if err := stream.Send(&workflow.WatchClusterEventsResponse{ClusterEvents: newEvents}); err != nil {
				return err
			}
		}
	}
}

// actionModelToClusterEvents derives cluster events from an Action model's phase.
func actionModelToClusterEvents(action *models.Action) []*workflow.ClusterEvent {
	phase := common.ActionPhase(action.Phase)
	if phase == common.ActionPhase_ACTION_PHASE_UNSPECIFIED {
		return nil
	}
	return []*workflow.ClusterEvent{{
		OccurredAt: timestamppb.New(action.UpdatedAt),
		Message:    phase.String(),
	}}
}

// actionModelToDetails converts a DB Action model to an ActionDetails proto.
func (s *RunService) actionModelToDetails(action *models.Action, actionID *common.ActionIdentifier) *workflow.ActionDetails {
	status := &workflow.ActionStatus{
		Phase:     common.ActionPhase(action.Phase),
		StartTime: timestamppb.New(action.CreatedAt),
		Attempts:  1,
	}
	if action.EndedAt.Valid {
		status.EndTime = timestamppb.New(action.EndedAt.Time)
	}

	phase := common.ActionPhase(action.Phase)
	attempt := &workflow.ActionAttempt{
		Attempt:   1,
		Phase:     phase,
		StartTime: timestamppb.New(action.CreatedAt),
	}
	if status.EndTime != nil {
		attempt.EndTime = status.EndTime
	}
	attempt.PhaseTransitions = []*workflow.PhaseTransition{{
		Phase:     phase,
		StartTime: timestamppb.New(action.CreatedAt),
		EndTime:   status.EndTime,
	}}

	return &workflow.ActionDetails{
		Id:       actionID,
		Status:   status,
		Attempts: []*workflow.ActionAttempt{attempt},
	}
}

// extractStorageURIs parses ActionSpec JSON to extract InputUri and RunOutputBase.
func extractStorageURIs(specJSON []byte) (inputURI, runOutputBase string) {
	if len(specJSON) == 0 {
		return
	}
	var spec struct {
		InputUri      string `json:"input_uri"`
		RunOutputBase string `json:"run_output_base"`
	}
	if err := json.Unmarshal(specJSON, &spec); err != nil {
		return
	}
	return spec.InputUri, spec.RunOutputBase
}

// Helper functions

// buildInputPrefix generates the input path prefix for the root action.
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

// buildRunOutputBase generates the output base path for the run.
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

	runID := &common.RunIdentifier{
		Org:     run.Org,
		Project: run.Project,
		Domain:  run.Domain,
		Name:    run.Name,
	}

	action := &workflow.Action{
		Id: &common.ActionIdentifier{
			Run:  runID,
			Name: run.Name,
		},
		Metadata: &workflow.ActionMetadata{},
		Status: &workflow.ActionStatus{
			Phase: common.ActionPhase(run.Phase),
		},
	}

	var actionDetails workflow.ActionDetails
	if err := json.Unmarshal(run.ActionDetails, &actionDetails); err == nil {
		action.Status.Attempts = actionDetails.Status.Attempts
		action.Status.CacheStatus = actionDetails.Status.CacheStatus
		if actionDetails.Status.StartTime != nil {
			action.Status.StartTime = actionDetails.Status.StartTime
		}
		if actionDetails.Status.EndTime != nil {
			action.Status.EndTime = actionDetails.Status.EndTime
		}
		if action.Status.StartTime != nil && action.Status.EndTime != nil {
			ms := uint64(action.Status.EndTime.AsTime().Sub(action.Status.StartTime.AsTime()).Milliseconds())
			action.Status.DurationMs = &ms
		}
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

	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     action.Org,
			Project: action.Project,
			Domain:  action.Domain,
			Name:    action.GetRunName(),
		},
		Name: action.Name,
	}

	actionStatus := &workflow.ActionStatus{
		Phase: common.ActionPhase(action.Phase),
	}

	var metadata *workflow.ActionMetadata
	if action.ParentActionName != nil {
		metadata = &workflow.ActionMetadata{
			Parent: *action.ParentActionName,
		}
	}

	return &workflow.EnrichedAction{
		Action: &workflow.Action{
			Id:       actionID,
			Metadata: metadata,
			Status:   actionStatus,
		},
		MeetsFilter: true,
	}
}

func (s *RunService) convertNodeUpdateToEnrichedProto(
	runID *common.RunIdentifier,
	update *nodeUpdate,
) *workflow.EnrichedAction {
	if update == nil || update.Node == nil || update.Node.Action == nil {
		return nil
	}

	action := update.Node.Action
	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     runID.Org,
			Project: runID.Project,
			Domain:  runID.Domain,
			Name:    runID.Name,
		},
		Name: action.Name,
	}

	actionStatus := &workflow.ActionStatus{
		Phase: common.ActionPhase(action.Phase),
	}

	var metadata *workflow.ActionMetadata
	if action.ParentActionName != nil {
		metadata = &workflow.ActionMetadata{
			Parent: *action.ParentActionName,
		}
	}

	childrenPhaseCounts := make(map[int32]int32, len(update.Node.ChildPhaseCounts))
	for phase, count := range update.Node.ChildPhaseCounts {
		childrenPhaseCounts[int32(phase)] = int32(count)
	}

	return &workflow.EnrichedAction{
		Action: &workflow.Action{
			Id:       actionID,
			Metadata: metadata,
			Status:   actionStatus,
		},
		MeetsFilter:         update.MeetsFilter,
		ChildrenPhaseCounts: childrenPhaseCounts,
	}
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

	rootActions, err := s.repo.ActionRepo().ListRootActions(ctx, org, project, domain, startDate, endDate, 1000)
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

	for _, action := range rootActions {
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

		phase := common.ActionPhase(action.Phase)
		g.phaseCounts[phase]++
		if phase == common.ActionPhase_ACTION_PHASE_FAILED {
			g.failCount++
		}

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
func extractShortName(name string) string {
	if name == "" {
		return ""
	}
	parts := strings.Split(name, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return name
}

// convertWatchRequestToListRequest converts a WatchRunsRequest to a ListRunsRequest
func (s *RunService) convertWatchRequestToListRequest(req *workflow.WatchRunsRequest) *workflow.ListRunsRequest {
	listReq := &workflow.ListRunsRequest{
		Request: &common.ListRequest{
			Limit: 100,
		},
	}

	switch target := req.Target.(type) {
	case *workflow.WatchRunsRequest_Org:
		listReq.ScopeBy = &workflow.ListRunsRequest_Org{
			Org: target.Org,
		}
	case *workflow.WatchRunsRequest_ProjectId:
		listReq.ScopeBy = &workflow.ListRunsRequest_ProjectId{
			ProjectId: target.ProjectId,
		}
	}

	return listReq
}

// runMatchesFilter checks if a run matches the WatchRunsRequest filter criteria
func (s *RunService) runMatchesFilter(run *models.Run, req *workflow.WatchRunsRequest) bool {
	if req.Target == nil {
		return true
	}

	switch target := req.Target.(type) {
	case *workflow.WatchRunsRequest_Org:
		return run.Org == target.Org
	case *workflow.WatchRunsRequest_ProjectId:
		return run.Org == target.ProjectId.Organization &&
			run.Project == target.ProjectId.Name &&
			run.Domain == target.ProjectId.Domain
	default:
		return true
	}
}
