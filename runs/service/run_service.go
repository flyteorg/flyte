package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/flyteorg/flyte/v2/app"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
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
	"github.com/flyteorg/flyte/v2/runs/repository/transformers"
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
	request := req.Msg
	// Validate request
	if err := request.Validate(); err != nil {
		logger.Errorf(ctx, "Invalid CreateRun request: %v", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Resolve run identity from request. When only a ProjectId is given, generate
	// the run name here and normalise the request to a RunId so that the repo
	// receives a single, pre-formed identifier — avoiding a second independent
	// name generation inside the repo layer.
	var runId *common.RunIdentifier
	switch id := request.Id.(type) {
	case *workflow.CreateRunRequest_RunId:
		runId = request.GetRunId()
	case *workflow.CreateRunRequest_ProjectId:
		runId = &common.RunIdentifier{
			Org:     id.ProjectId.Organization,
			Project: id.ProjectId.Name,
			Domain:  id.ProjectId.Domain,
			Name:    generateRunName(time.Now().UnixNano()),
		}
		request.Id = &workflow.CreateRunRequest_RunId{
			RunId: runId,
		}
	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid run ID type"))
	}

	actionID := &common.ActionIdentifier{
		Run:  runId,
		Name: runId.Name,
	}

	// Get the task template and taskID
	var taskID *task.TaskIdentifier
	var taskSpec *task.TaskSpec
	var err error
	inputs := request.GetInputs()
	runSpec := request.GetRunSpec()
	// TODO: Add trigger
	switch request.GetTask().(type) {
	case *workflow.CreateRunRequest_TaskId:
		taskID = request.GetTaskId()
		taskSpec, err = fetchTaskSpecByID(ctx, s.repo.TaskRepo(), taskID)

		if err != nil {
			return nil, err
		}
	case *workflow.CreateRunRequest_TaskSpec:
		taskSpec = request.GetTaskSpec()
		taskID = taskIdFromTaskSpec(taskSpec)
	}

	if runSpec == nil {
		runSpec = &task.RunSpec{}
	}
	request.RunSpec = runSpec

	inputs = fillDefaultInputs(inputs, taskSpec.GetDefaultInputs())
	// Compute storage URIs before DB insert so they're persisted in the ActionSpec
	inputPrefix := buildInputPrefix(s.storagePrefix, runId.GetOrg(), runId.GetProject(), runId.GetDomain(), runId.GetName())
	runOutputBase := buildRunOutputBase(s.storagePrefix, runId.GetOrg(), runId.GetProject(), runId.GetDomain(), runId.GetName())
	if runSpec.GetRawDataStorage() == nil || runSpec.GetRawDataStorage().GetRawDataPrefix() == "" {
		runSpec.RawDataStorage = &task.RawDataStorage{RawDataPrefix: s.storagePrefix}
	}

	// Persist the full Inputs proto so context survives CreateRun -> storage -> runtime.
	inputRef := storage.DataReference(inputPrefix + "/inputs.pb")
	if err := s.dataStore.WriteProtobuf(ctx, inputRef, storage.Options{}, inputs); err != nil {
		logger.Errorf(ctx, "Failed to write inputs to storage: %v", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to write inputs: %w", err))
	}
	logger.Infof(ctx, "Wrote inputs to %s", inputRef)

	cacheKey := ""
	// Generate cache key for the root action if its task is discoverable
	// Executor determines if an action is cacheable on the setting of the cache key
	if taskSpec.GetTaskTemplate().GetMetadata().GetDiscoverable() {
		cacheKey, err = generateCacheKeyForTask(taskSpec.GetTaskTemplate(), inputs)
		if err != nil {
			logger.Warnf(ctx, "Failed to generate cache key for root action %v: %v", actionID, err)
		}
	}

	// Create a run in a database with storage URIs
	run, err := s.repo.ActionRepo().CreateRun(ctx, request, inputPrefix, runOutputBase)
	if err != nil {
		logger.Errorf(ctx, "Failed to create run: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	_, err = s.actionsClient.Enqueue(ctx, connect.NewRequest(&actions.EnqueueRequest{
		Action: &actions.Action{
			ActionId:      actionID,
			InputUri:      inputPrefix,
			RunOutputBase: runOutputBase,
			Spec: &actions.Action_Task{
				Task: &workflow.TaskAction{
					Id:       taskID,
					Spec:     taskSpec,
					CacheKey: wrapperspb.String(cacheKey),
				},
			},
		},
		RunSpec: runSpec,
	}))
	if err != nil {
		logger.Errorf(ctx, "Failed to enqueue root action: %v", err)
	} else {
		logger.Infof(ctx, "Successfully enqueued root action for run %s", runId.Name)
	}

	return connect.NewResponse(&workflow.CreateRunResponse{
		Run: s.convertRunToProto(run),
	}), nil
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

	actionDetails, err := s.buildActionDetails(ctx, run, rootActionID)
	if err != nil {
		logger.Errorf(ctx, "Failed to build run action details: %v", err)
		return nil, err
	}

	details := &workflow.RunDetails{
		RunSpec: extractRunSpec(run.ActionSpec),
		Action:  actionDetails,
	}

	logger.Infof(ctx, "Retrieved run details for: %s", run.Name)
	return connect.NewResponse(&workflow.GetRunDetailsResponse{Details: details}), nil
}

// GetActionDetails gets detailed information about an action from the DB.
func (s *RunService) GetActionDetails(
	ctx context.Context,
	req *connect.Request[workflow.GetActionDetailsRequest],
) (*connect.Response[workflow.GetActionDetailsResponse], error) {
	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	details, err := s.getActionDetails(ctx, req.Msg.GetActionId())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&workflow.GetActionDetailsResponse{Details: details}), nil
}

func (s *RunService) getActionDetails(ctx context.Context, actionId *common.ActionIdentifier) (*workflow.ActionDetails, error) {
	model, err := s.repo.ActionRepo().GetAction(ctx, actionId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("action not found: %w", err))
	}
	return s.buildActionDetails(ctx, model, actionId)
}

// buildActionDetails enriches a pre-fetched action model with task spec and attempts.
func (s *RunService) buildActionDetails(ctx context.Context, model *models.Action, actionId *common.ActionIdentifier) (*workflow.ActionDetails, error) {
	// Get task spec and attempts in parallel
	var action *workflow.ActionDetails
	var eg errgroup.Group
	eg.Go(func() error {
		var info *workflow.RunInfo
		if len(model.DetailedInfo) > 0 {
			info = &workflow.RunInfo{}
			if err := proto.Unmarshal(model.DetailedInfo, info); err != nil {
				return err
			}
		}

		action = s.actionModelToDetails(model, actionId)

		// Get task spec by digest if available
		if info.GetTaskSpecDigest() != "" {
			specModel, err := s.repo.TaskRepo().GetTaskSpec(ctx, info.GetTaskSpecDigest())
			if err != nil {
				if ctx.Err() == nil {
					logger.Errorf(ctx, "failed to get task spec for action %v: %v", actionId, err)
				}
				return err
			}

			// Fill in the action spec based on the action type
			switch action.GetMetadata().GetActionType() {
			case workflow.ActionType_ACTION_TYPE_UNSPECIFIED:
				logger.Warnf(ctx, "action %v has unspecified action type, defaulting to task", actionId)
			case workflow.ActionType_ACTION_TYPE_TASK:
				spec, err := transformers.ToTaskSpec(specModel)
				if err != nil {
					logger.Errorf(ctx, "failed to convert task spec model for action %v: %v", actionId, err)
					return err
				}
				action.Spec = &workflow.ActionDetails_Task{Task: spec}
			case workflow.ActionType_ACTION_TYPE_TRACE:
				spec, err := transformers.ToTraceSpec(specModel)
				if err != nil {
					logger.Errorf(ctx, "failed to convert trace spec model for action %v: %v", actionId, err)
					return err
				}
				action.Spec = &workflow.ActionDetails_Trace{Trace: spec}
			default:
				return connect.NewError(connect.CodeInternal,
					fmt.Errorf("unknown action type %v for action %v", action.GetMetadata().GetActionType(), actionId))
			}
		}

		if action.Spec == nil {
			// Fall back to the embedded ActionSpec when no spec was loaded from task table.
			setActionDetailsSpecFromActionSpec(action, model.ActionSpec)
		}

		return nil
	})

	// Get attempts
	var attempts []*workflow.ActionAttempt
	eg.Go(func() error {
		var err error
		attempts, err = s.getAttempts(ctx, actionId)
		if err != nil {
			if ctx.Err() == nil {
				logger.Warnf(ctx, "failed to get action attempts for action %v: %v", actionId, err)
			}
			return err
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		if ctx.Err() == nil {
			logger.Errorf(ctx, "failed to get action details for action %v: %v", actionId, err)
		}
		return nil, err
	}

	action.Attempts = attempts

	switch action.GetStatus().GetPhase() {
	case common.ActionPhase_ACTION_PHASE_FAILED:
		// Get action error from last attempt. Events are eventually consistent, so we may not have
		// information from the latest attempt yet.
		numAttempts := len(action.GetAttempts())
		if numAttempts > 0 && action.GetAttempts()[numAttempts-1].GetAttempt() == action.GetStatus().GetAttempts() {
			action.Result = &workflow.ActionDetails_ErrorInfo{
				ErrorInfo: action.GetAttempts()[numAttempts-1].GetErrorInfo(),
			}
		}
	case common.ActionPhase_ACTION_PHASE_ABORTED:
		// TODO: set AbortInfo
	}

	return action, nil
}

// getAttempts looks up all events for this action and builds a list of attempts.
func (s *RunService) getAttempts(ctx context.Context, actionId *common.ActionIdentifier) ([]*workflow.ActionAttempt, error) {
	eventModels, err := s.repo.ActionRepo().ListEvents(ctx, actionId, 500)
	if err != nil {
		return nil, err
	}

	events := make([]*workflow.ActionEvent, 0, len(eventModels))
	for _, m := range eventModels {
		event, err := m.ToActionEvent()
		if err != nil {
			logger.Warnf(ctx, "failed to convert action event model for action %v: %v", actionId, err)
			return nil, err
		}
		events = append(events, event)
	}

	// Group events by attempt
	attemptToEvents := map[uint32][]*workflow.ActionEvent{}
	for _, event := range events {
		attemptToEvents[event.GetAttempt()] = append(attemptToEvents[event.GetAttempt()], event)
	}

	attempts := make([]*workflow.ActionAttempt, 0, len(attemptToEvents))
	for attempt, evts := range attemptToEvents {
		attempts = append(attempts, mergeEvents(attempt, evts))
	}
	sort.SliceStable(attempts, func(i, j int) bool {
		return attempts[i].GetAttempt() < attempts[j].GetAttempt()
	})

	return attempts, nil
}

// mergeEvents merges a set of events for the same attempt into a single ActionAttempt.
func mergeEvents(attempt uint32, events []*workflow.ActionEvent) *workflow.ActionAttempt {
	// Order events by the controller-observed phase transition time.
	// ReportedTime reflects when an observation was emitted and can arrive out of order,
	// so it is only used as a tie-breaker for otherwise-identical UpdatedTime values.
	sort.SliceStable(events, func(i, j int) bool {
		updatedTimeI := events[i].GetUpdatedTime()
		updatedTimeJ := events[j].GetUpdatedTime()
		if updatedTimeI != nil && updatedTimeJ != nil && !updatedTimeI.AsTime().Equal(updatedTimeJ.AsTime()) {
			return updatedTimeI.AsTime().Before(updatedTimeJ.AsTime())
		}

		reportedTimeI := events[i].GetReportedTime()
		reportedTimeJ := events[j].GetReportedTime()
		if reportedTimeI != nil && reportedTimeJ != nil && !reportedTimeI.AsTime().Equal(reportedTimeJ.AsTime()) {
			return reportedTimeI.AsTime().Before(reportedTimeJ.AsTime())
		}

		return phaseOrder(events[i].GetPhase()) < phaseOrder(events[j].GetPhase())
	})

	if len(events) == 0 {
		return &workflow.ActionAttempt{Attempt: attempt}
	}

	// Merge log info (latest takes precedence).
	logs := map[string]*core.TaskLog{}
	streamingLogsAvailable := false
	reportURI := ""
	lastEvent := events[len(events)-1]
	var logContext *core.LogContext
	for i := len(events) - 1; i >= 0; i-- {
		event := events[i]
		if event.GetLogContext() != nil && logContext == nil {
			logContext = event.GetLogContext()
			streamingLogsAvailable = true
		}

		for _, log := range event.GetLogInfo() {
			if _, ok := logs[log.GetName()]; !ok {
				logs[log.GetName()] = log
			}
		}
		if reportURI == "" && event.GetOutputs().GetReportUri() != "" {
			reportURI = event.GetOutputs().GetReportUri()
		}
		phase := events[i].GetPhase()
		if IsTerminalPhase(phase) {
			lastEvent = events[i]
		}
	}

	sortedLogs := make([]*core.TaskLog, 0, len(logs))
	for _, log := range logs {
		sortedLogs = append(sortedLogs, log)
	}
	sort.SliceStable(sortedLogs, func(i, j int) bool {
		return sortedLogs[i].GetName() < sortedLogs[j].GetName()
	})

	if reportURI != "" {
		if lastEvent.GetOutputs() == nil {
			lastEvent.Outputs = &task.OutputReferences{}
		}
		lastEvent.Outputs.ReportUri = reportURI
	}

	startTime := events[0].GetUpdatedTime()
	var endTime *timestamppb.Timestamp
	if IsTerminalPhase(lastEvent.GetPhase()) {
		endTime = lastEvent.GetUpdatedTime()
	}
	// Ensure endTime is never before startTime due to timestamp granularity differences.
	if endTime != nil && startTime.AsTime().After(endTime.AsTime()) {
		endTime = startTime
	}

	var clusterEvents []*workflow.ClusterEvent
	for _, event := range events {
		clusterEvents = append(clusterEvents, event.GetClusterEvents()...)
	}

	// Build phase transitions
	phaseTransitions := make([]*workflow.PhaseTransition, 0, len(common.ActionPhase_name))
	phaseSeen := map[common.ActionPhase]bool{}
	for _, event := range events {
		phase := event.GetPhase()
		if !phaseSeen[phase] {
			if len(phaseTransitions) > 0 {
				phaseTransitions[len(phaseTransitions)-1].EndTime = event.GetUpdatedTime()
			}
			phaseTransitions = append(phaseTransitions, &workflow.PhaseTransition{
				Phase:     phase,
				StartTime: event.GetUpdatedTime(),
			})
			if IsTerminalPhase(phase) {
				phaseTransitions[len(phaseTransitions)-1].EndTime = event.GetUpdatedTime()
				break
			}
			phaseSeen[phase] = true
		}
	}

	return &workflow.ActionAttempt{
		Attempt:          attempt,
		Phase:            lastEvent.GetPhase(),
		StartTime:        startTime,
		EndTime:          endTime,
		Outputs:          lastEvent.GetOutputs(),
		ErrorInfo:        lastEvent.GetErrorInfo(),
		LogsAvailable:    streamingLogsAvailable,
		LogContext:       logContext,
		LogInfo:          sortedLogs,
		CacheStatus:      lastEvent.GetCacheStatus(),
		ClusterEvents:    clusterEvents,
		Cluster:          lastEvent.GetCluster(),
		PhaseTransitions: phaseTransitions,
	}
}

func phaseOrder(phase common.ActionPhase) int {
	switch phase {
	case common.ActionPhase_ACTION_PHASE_QUEUED:
		return 0
	case common.ActionPhase_ACTION_PHASE_WAITING_FOR_RESOURCES:
		return 1
	case common.ActionPhase_ACTION_PHASE_INITIALIZING:
		return 2
	case common.ActionPhase_ACTION_PHASE_RUNNING:
		return 3
	case common.ActionPhase_ACTION_PHASE_SUCCEEDED,
		common.ActionPhase_ACTION_PHASE_FAILED,
		common.ActionPhase_ACTION_PHASE_ABORTED,
		common.ActionPhase_ACTION_PHASE_TIMED_OUT:
		return 4
	default:
		return 5
	}
}

func IsTerminalPhase(phase common.ActionPhase) bool {
	return phase == common.ActionPhase_ACTION_PHASE_FAILED ||
		phase == common.ActionPhase_ACTION_PHASE_SUCCEEDED ||
		phase == common.ActionPhase_ACTION_PHASE_TIMED_OUT ||
		phase == common.ActionPhase_ACTION_PHASE_ABORTED
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

	inputURI, _ := extractStorageURIs(action.ActionSpec)

	info := &workflow.RunInfo{}
	if err := proto.Unmarshal(action.DetailedInfo, info); err != nil {
		return nil, err
	}

	resp := &workflow.GetActionDataResponse{
		Inputs:  &task.Inputs{},
		Outputs: &task.Outputs{},
	}

	// Read inputs from storage
	group, groupCtx := errgroup.WithContext(ctx)
	if inputURI != "" {
		group.Go(func() error {
			inputRef := storage.DataReference(inputURI)
			logger.Debugf(groupCtx, "Reading inputs from: %s", inputRef)
			if err := s.dataStore.ReadProtobuf(groupCtx, inputRef, resp.Inputs); err != nil {
				if !storage.IsNotFound(err) {
					logger.Errorf(groupCtx, "Failed to read inputs from %s: %v", inputRef, err)
					return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to read inputs: %w", err))
				}
				logger.Debugf(groupCtx, "Inputs not found at %s", inputRef)
			} else {
				logger.Debugf(groupCtx, "Read %d input literals and %d action contexts", len(resp.Inputs.Literals), len(resp.Inputs.Context))
			}
			return nil
		})
	} else {
		logger.Warnf(ctx, "Action %s has empty InputURI", req.Msg.ActionId.Name)
	}

	// Read outputs from storage (only present if action succeeded)
	if action.Phase == int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED) {
		group.Go(func() error {
			// There are no attempts for trace actions, so we can skip the attempt validation
			var attempts []*workflow.ActionAttempt
			var err error
			if workflow.ActionType(action.ActionType) == workflow.ActionType_ACTION_TYPE_TRACE {
				if info.GetOutputsUri() == "" {
					return nil
				}
				logger.Debugf(groupCtx, "Reading outputs from: %s", info.GetOutputsUri())

				outputMap := &core.LiteralMap{}
				if err := s.dataStore.ReadProtobuf(groupCtx, storage.DataReference(info.GetOutputsUri()), outputMap); err != nil {
					if !storage.IsNotFound(err) {
						logger.Errorf(groupCtx, "Failed to read outputs from %s: %v", info.GetOutputsUri(), err)
						return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to read outputs: %w", err))
					}
					logger.Debugf(groupCtx, "Outputs not found at %s (action may not have finished)", info.GetOutputsUri())
				} else {
					resp.Outputs = literalMapToOutputs(outputMap)
					logger.Debugf(groupCtx, "Read %d output literals", len(resp.Outputs.Literals))
				}

				return nil
			}

			// Default to "task" action types
			attempts, err = s.getAttempts(groupCtx, req.Msg.GetActionId())
			if err != nil {
				return err
			}

			if len(attempts) == 0 {
				return app.NewServerError(codes.NotFound, "outputs not available, no attempts for action")
			}

			outputUri := attempts[len(attempts)-1].GetOutputs().GetOutputUri()
			if outputUri == "" {
				return app.NewServerError(codes.NotFound, "outputs not available")
			}

			logger.Debugf(groupCtx, "Reading outputs from: %s", outputUri)
			outputMap := &core.LiteralMap{}
			if err := s.dataStore.ReadProtobuf(groupCtx, storage.DataReference(outputUri), outputMap); err != nil {
				if !storage.IsNotFound(err) {
					logger.Errorf(groupCtx, "Failed to read outputs from %s: %v", outputUri, err)
					return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to read outputs: %w", err))
				}
				logger.Debugf(groupCtx, "Outputs not found at %s (action may not have finished)", outputUri)
			} else {
				resp.Outputs = literalMapToOutputs(outputMap)
				logger.Debugf(groupCtx, "Read %d output literals", len(resp.Outputs.Literals))
			}

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return nil, err
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
	updates := make(chan *models.Run, 50)
	errs := make(chan error, 1)

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
	details, err := s.getActionDetails(ctx, req.Msg.GetActionId())
	if err != nil {
		return connect.NewError(connect.CodeNotFound, fmt.Errorf("failed to get action details: %w", err))
	}

	if err := stream.Send(&workflow.WatchActionDetailsResponse{
		Details: details,
	}); err != nil {
		return err
	}

	// If the action is already in a terminal phase, no further updates are expected.
	if IsTerminalPhase(details.GetStatus().GetPhase()) {
		return nil
	}

	// Step 2: Watch DB for updates
	updates := make(chan *models.Action, 50)
	errs := make(chan error, 1)
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
			// Reuse the already-fetched action model from WatchActionUpdates
			details, err := s.buildActionDetails(ctx, updated, actionID)
			if err != nil {
				if ctx.Err() != nil {
					return nil
				}
				logger.Errorf(ctx, "failed to get action details for action %s: %v", actionID.Name, err)
				return connect.NewError(connect.CodeInternal, err)
			}
			if err := stream.Send(&workflow.WatchActionDetailsResponse{
				Details: details,
			}); err != nil {
				return err
			}
			// Close the stream once the action reaches a terminal phase.
			if IsTerminalPhase(details.GetStatus().GetPhase()) {
				return nil
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
	updatesCh := make(chan *models.Run, 50)
	errsCh := make(chan error, 1)
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
	updatesCh := make(chan *models.Action, 50)
	errsCh := make(chan error, 1)
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

	// If the action is already in a terminal phase, no further events are expected.
	if IsTerminalPhase(common.ActionPhase(action.Phase)) {
		return nil
	}

	// Step 2: Watch for updates from DB
	updatesCh := make(chan *models.Action, 50)
	errsCh := make(chan error, 1)
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
			if len(newEvents) > 0 {
				if err := stream.Send(&workflow.WatchClusterEventsResponse{ClusterEvents: newEvents}); err != nil {
					return err
				}
			}
			// Close the stream once the action reaches a terminal phase.
			if IsTerminalPhase(common.ActionPhase(updated.Phase)) {
				return nil
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
		Phase:       common.ActionPhase(action.Phase),
		Attempts:    action.Attempts,
		CacheStatus: action.CacheStatus,
	}
	if action.StartedAt.Valid {
		status.StartTime = timestamppb.New(action.StartedAt.Time)
	}
	if action.EndedAt.Valid {
		status.EndTime = timestamppb.New(action.EndedAt.Time)
	}
	if action.DurationMs.Valid {
		durationMs := uint64(action.DurationMs.Int64)
		status.DurationMs = &durationMs
	}

	metadata := actionMetadataFromModel(action)

	return &workflow.ActionDetails{
		Id:       actionID,
		Metadata: metadata,
		Status:   status,
	}
}

func setActionDetailsSpecFromActionSpec(details *workflow.ActionDetails, actionSpecBytes []byte) {
	specMsg := extractActionSpec(actionSpecBytes)
	if specMsg == nil {
		return
	}

	switch s := specMsg.Spec.(type) {
	case *workflow.ActionSpec_Task:
		if s.Task.GetSpec() != nil {
			details.Spec = &workflow.ActionDetails_Task{Task: s.Task.GetSpec()}
		}
	case *workflow.ActionSpec_Trace:
		if s.Trace.GetSpec() != nil {
			details.Spec = &workflow.ActionDetails_Trace{Trace: s.Trace.GetSpec()}
		}
	}
}

// actionMetadataFromModel builds an ActionMetadata proto from DB model columns.
func actionMetadataFromModel(action *models.Action) *workflow.ActionMetadata {
	metadata := &workflow.ActionMetadata{
		ActionType:  workflow.ActionType(action.ActionType),
		FuntionName: action.FunctionName,
	}
	if action.ParentActionName.Valid {
		metadata.Parent = action.ParentActionName.String
	}
	if action.ActionGroup.Valid {
		metadata.Group = action.ActionGroup.String
	}
	if action.EnvironmentName.Valid {
		metadata.EnvironmentName = action.EnvironmentName.String
	}

	switch workflow.ActionType(action.ActionType) {
	case workflow.ActionType_ACTION_TYPE_TASK:
		taskMeta := &workflow.TaskActionMetadata{
			TaskType: action.TaskType,
		}
		if action.TaskName.Valid {
			taskMeta.Id = &task.TaskIdentifier{
				Org:     action.TaskOrg.String,
				Project: action.TaskProject.String,
				Domain:  action.TaskDomain.String,
				Name:    action.TaskName.String,
				Version: action.TaskVersion.String,
			}
		}
		if action.TaskShortName.Valid {
			taskMeta.ShortName = action.TaskShortName.String
		}
		metadata.Spec = &workflow.ActionMetadata_Task{Task: taskMeta}
	case workflow.ActionType_ACTION_TYPE_TRACE:
		traceMeta := &workflow.TraceActionMetadata{
			Name: action.FunctionName,
		}
		metadata.Spec = &workflow.ActionMetadata_Trace{Trace: traceMeta}
	}

	return metadata
}

// extractStorageURIs parses ActionSpec protobuf to extract InputUri and RunOutputBase.
func extractStorageURIs(specBytes []byte) (inputURI, runOutputBase string) {
	if len(specBytes) == 0 {
		return
	}
	var spec workflow.ActionSpec
	if err := proto.Unmarshal(specBytes, &spec); err != nil {
		return
	}
	return spec.GetInputUri(), spec.GetRunOutputBase()
}

// extractRunSpec parses ActionSpec JSON to extract the RunSpec
func extractActionSpec(specProto []byte) *workflow.ActionSpec {
	if len(specProto) == 0 {
		return nil
	}
	var specMsg workflow.ActionSpec
	if err := json.Unmarshal(specProto, &specMsg); err != nil {
		return nil
	}
	return &specMsg
}

// extractRunSpec parses ActionSpec JSON to extract the RunSpec
func extractRunSpec(specProto []byte) *task.RunSpec {
	if len(specProto) == 0 {
		return nil
	}
	spec := extractActionSpec(specProto)
	if spec == nil {
		return nil
	}
	return spec.RunSpec
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
	if inputs == nil || len(inputs.Literals) == 0 {
		return &core.LiteralMap{Literals: map[string]*core.Literal{}}
	}
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
		Metadata: actionMetadataFromModel(run),
		Status: &workflow.ActionStatus{
			Phase:       common.ActionPhase(run.Phase),
			Attempts:    run.Attempts,
			CacheStatus: run.CacheStatus,
		},
	}
	if run.EndedAt.Valid {
		action.Status.EndTime = timestamppb.New(run.EndedAt.Time)
	}
	if run.DurationMs.Valid && run.DurationMs.Int64 >= 0 {
		ms := uint64(run.DurationMs.Int64)
		action.Status.DurationMs = &ms
	}

	var actionDetails workflow.ActionDetails
	if err := json.Unmarshal(run.ActionDetails, &actionDetails); err != nil {
		if len(run.ActionDetails) > 0 {
			logger.Warningf(context.Background(), "failed to unmarshal action details for run %s: %v", run.Name, err)
		}
	} else if actionDetails.Status != nil {
		action.Status.Attempts = actionDetails.Status.Attempts
		action.Status.CacheStatus = actionDetails.Status.CacheStatus
		if actionDetails.Status.StartTime != nil {
			action.Status.StartTime = actionDetails.Status.StartTime
		}
		if actionDetails.Status.EndTime != nil {
			action.Status.EndTime = actionDetails.Status.EndTime
		}
		if actionDetails.Status.DurationMs != nil {
			action.Status.DurationMs = actionDetails.Status.DurationMs
		} else if action.Status.StartTime != nil && action.Status.EndTime != nil {
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
		Phase:       common.ActionPhase(action.Phase),
		Attempts:    action.Attempts,
		CacheStatus: action.CacheStatus,
	}

	if action.StartedAt.Valid {
		actionStatus.StartTime = timestamppb.New(action.StartedAt.Time)
	}

	if action.EndedAt.Valid {
		actionStatus.EndTime = timestamppb.New(action.EndedAt.Time)
	}

	if action.DurationMs.Valid && action.DurationMs.Int64 > 0 {
		durationMs := uint64(action.DurationMs.Int64)
		actionStatus.DurationMs = &durationMs
	}

	var metadata *workflow.ActionMetadata
	if action.ParentActionName.Valid {
		metadata = &workflow.ActionMetadata{
			Parent: action.ParentActionName.String,
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
		Phase:       common.ActionPhase(action.Phase),
		Attempts:    action.Attempts,
		CacheStatus: action.CacheStatus,
	}

	if action.StartedAt.Valid {
		actionStatus.StartTime = timestamppb.New(action.StartedAt.Time)
	}

	if action.EndedAt.Valid {
		actionStatus.EndTime = timestamppb.New(action.EndedAt.Time)
	}

	if action.DurationMs.Valid && action.DurationMs.Int64 > 0 {
		durationMs := uint64(action.DurationMs.Int64)
		actionStatus.DurationMs = &durationMs
	}

	metadata := &workflow.ActionMetadata{
		Parent: CoalesceNullString(action.ParentActionName),
		Group:  CoalesceNullString(action.ActionGroup),
	}

	// Pivot on known task types for response types
	switch workflow.ActionType(action.ActionType) {
	case workflow.ActionType_ACTION_TYPE_TRACE:
		metadata.Spec = &workflow.ActionMetadata_Trace{
			Trace: &workflow.TraceActionMetadata{
				Name: CoalesceNullString(action.TaskName),
			},
		}
	case workflow.ActionType_ACTION_TYPE_TASK:
		metadata.Spec = &workflow.ActionMetadata_Task{
			Task: &workflow.TaskActionMetadata{
				Id: &task.TaskIdentifier{
					Org:     CoalesceNullString(action.TaskOrg),
					Project: CoalesceNullString(action.TaskProject),
					Domain:  CoalesceNullString(action.TaskDomain),
					Name:    CoalesceNullString(action.TaskName),
					Version: CoalesceNullString(action.TaskVersion),
				},
				TaskType: action.TaskType,
			},
		}
		// Check if there's a task short name override, if so, use that. Otherwise, fall back to the parsed task name.
		if action.TaskShortName.Valid {
			metadata.GetTask().ShortName = action.TaskShortName.String
		} else {
			metadata.GetTask().ShortName = transformers.ExtractFunctionName(context.Background(), CoalesceNullString(action.TaskName), "")
		}
		metadata.EnvironmentName = action.EnvironmentName.String

	case workflow.ActionType_ACTION_TYPE_CONDITION:
		// Unhandled for now
	case workflow.ActionType_ACTION_TYPE_UNSPECIFIED:
		// No-op, this should never happen.
	}

	metadata.FuntionName = action.FunctionName
	// TODO: Add ExecutedBy, TriggerTaskName, TriggerId

	metadata.Source = workflow.RunSource(workflow.RunSource_value[action.RunSource])

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

// extractTaskName attempts to extract the task name from an ActionSpec protobuf blob.
func extractTaskName(specBytes []byte) string {
	if len(specBytes) == 0 {
		return ""
	}
	var spec workflow.ActionSpec
	if err := proto.Unmarshal(specBytes, &spec); err != nil {
		return ""
	}
	if taskAction := spec.GetTask(); taskAction != nil {
		if id := taskAction.GetId(); id != nil && id.GetName() != "" {
			return id.GetName()
		}
		if taskSpec := taskAction.GetSpec(); taskSpec != nil && taskSpec.GetShortName() != "" {
			return taskSpec.GetShortName()
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

func fetchTaskSpecByID(ctx context.Context, taskRepo interfaces.TaskRepo, taskID *task.TaskIdentifier) (*task.TaskSpec, error) {
	taskModel, err := taskRepo.GetTask(ctx, transformers.ToTaskKey(taskID))
	if err != nil {
		return nil, err
	}
	tasks, err := transformers.TaskModelsToTaskDetailsWithoutIdentity(ctx, []*models.Task{taskModel})
	if err != nil {
		return nil, err
	}
	return tasks[0].GetSpec(), nil
}

func fillDefaultInputs(inputs *task.Inputs, defaultInputs []*task.NamedParameter) *task.Inputs {
	inputMap := make(map[string]struct{}, len(inputs.GetLiterals()))
	for _, literal := range inputs.GetLiterals() {
		inputMap[literal.GetName()] = struct{}{}
	}
	if inputs == nil {
		inputs = &task.Inputs{}
	}
	for _, param := range defaultInputs {
		if _, ok := inputMap[param.GetName()]; !ok {
			defaultValue := param.GetParameter().GetDefault()
			if defaultValue != nil {
				inputs.Literals = append(inputs.Literals, &task.NamedLiteral{
					Name:  param.GetName(),
					Value: defaultValue,
				})
			}
		}
	}
	return inputs
}

func taskIdFromTaskSpec(taskSpec *task.TaskSpec) *task.TaskIdentifier {
	if taskSpec == nil {
		return nil
	}
	return &task.TaskIdentifier{
		Org:     taskSpec.GetTaskTemplate().GetId().GetOrg(),
		Project: taskSpec.GetTaskTemplate().GetId().GetProject(),
		Domain:  taskSpec.GetTaskTemplate().GetId().GetDomain(),
		Name:    taskSpec.GetTaskTemplate().GetId().GetName(),
		Version: taskSpec.GetTaskTemplate().GetId().GetVersion(),
	}
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
