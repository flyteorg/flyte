package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/sync/errgroup"
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
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy/dataproxyconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/project/projectconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	"github.com/flyteorg/flyte/v2/runs/repository/impl"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
	"github.com/flyteorg/flyte/v2/runs/repository/transformers"
)

// RunService implements the RunServiceHandler interface
type RunService struct {
	repo            interfaces.Repository
	actionsClient   actionsconnect.ActionsServiceClient
	dataProxyClient actionDataClient
	projectClient   projectconnect.ProjectServiceClient
	storagePrefix   string
	dataStore       *storage.DataStore
	abortReconciler *AbortReconciler
}

type actionDataClient interface {
	GetActionData(
		ctx context.Context,
		req *connect.Request[dataproxy.GetActionDataRequest],
	) (*connect.Response[dataproxy.GetActionDataResponse], error)
}

const (
	runIDLength     = 20
	runStringFormat = "r%s"
	RootActionName  = "a0"
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
func NewRunService(
	repo interfaces.Repository,
	actionsClient actionsconnect.ActionsServiceClient,
	dataProxyClient dataproxyconnect.DataProxyServiceClient,
	projectClient projectconnect.ProjectServiceClient,
	storagePrefix string,
	dataStore *storage.DataStore,
	reconciler *AbortReconciler,
) *RunService {
	return &RunService{
		repo:            repo,
		actionsClient:   actionsClient,
		dataProxyClient: dataProxyClient,
		projectClient:   projectClient,
		storagePrefix:   storagePrefix,
		dataStore:       dataStore,
		abortReconciler: reconciler,
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

	if err := validateProjectExists(ctx, s.projectClient, runId.GetProject()); err != nil {
		return nil, err
	}

	actionID := &common.ActionIdentifier{
		Run:  runId,
		Name: RootActionName,
	}

	// Get the task template and taskID
	var taskID *task.TaskIdentifier
	var taskSpec *task.TaskSpec
	var triggerName, triggerTaskName, triggerType string
	var triggerRevision int64
	var err error
	runSpec := request.GetRunSpec()
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
	case *workflow.CreateRunRequest_TriggerName:
		tn := request.GetTriggerName()
		triggerModel, triggerErr := s.repo.TriggerRepo().GetTrigger(ctx, transformers.ToTriggerKey(tn))
		if triggerErr != nil {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("trigger %q not found: %w", tn.GetName(), triggerErr))
		}
		triggerName = tn.GetName()
		triggerTaskName = triggerModel.TaskName
		triggerRevision = int64(triggerModel.LatestRevision)
		triggerType = triggerModel.AutomationType
		taskID = &task.TaskIdentifier{
			Project: triggerModel.Project,
			Domain:  triggerModel.Domain,
			Name:    triggerModel.TaskName,
			Version: triggerModel.TaskVersion,
		}
		taskSpec, err = fetchTaskSpecByID(ctx, s.repo.TaskRepo(), taskID)
		if err != nil {
			return nil, err
		}
	}

	if runSpec == nil {
		runSpec = &task.RunSpec{}
	}
	request.RunSpec = runSpec

	// Compute storage URIs before DB insert so they're persisted in the ActionSpec
	inputPrefix := buildInputPrefix(s.storagePrefix, runId.GetProject(), runId.GetDomain(), runId.GetName())
	runOutputBase := buildRunOutputBase(s.storagePrefix, runId.GetProject(), runId.GetDomain(), runId.GetName())
	if runSpec.GetRawDataStorage() == nil || runSpec.GetRawDataStorage().GetRawDataPrefix() == "" {
		runSpec.RawDataStorage = &task.RawDataStorage{RawDataPrefix: s.storagePrefix}
	}

	cacheKey := ""
	switch iw := request.GetInputWrapper().(type) {
	case *workflow.CreateRunRequest_OffloadedInputData:
		// Inputs are already persisted by UploadInputs. Use the offloaded URI directly.
		inputPrefix = iw.OffloadedInputData.GetUri()

		if taskSpec.GetTaskTemplate().GetMetadata().GetDiscoverable() {
			cacheKey, err = generateCacheKeyForTask(taskSpec.GetTaskTemplate(), iw.OffloadedInputData.GetInputsHash())
			if err != nil {
				logger.Warnf(ctx, "Failed to generate cache key for root action %v: %v", actionID, err)
			}
		}
	default:
		inputs := request.GetInputs()
		inputs = fillDefaultInputs(inputs, taskSpec.GetDefaultInputs())

		// Persist the full Inputs proto so context survives CreateRun -> storage -> runtime.
		inputRef := storage.DataReference(inputPrefix + "/inputs.pb")
		if err := s.dataStore.WriteProtobuf(ctx, inputRef, storage.Options{}, inputs); err != nil {
			logger.Errorf(ctx, "Failed to write inputs to storage: %v", err)
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to write inputs: %w", err))
		}
		logger.Infof(ctx, "Wrote inputs to %s", inputRef)

		if taskSpec.GetTaskTemplate().GetMetadata().GetDiscoverable() {
			inputsHash, hashErr := computeFilteredInputsHash(taskSpec.GetTaskTemplate(), inputs)
			if hashErr != nil {
				logger.Warnf(ctx, "Failed to hash inputs for root action %v: %v", actionID, hashErr)
			} else {
				cacheKey, err = generateCacheKeyForTask(taskSpec.GetTaskTemplate(), inputsHash)
				if err != nil {
					logger.Warnf(ctx, "Failed to generate cache key for root action %v: %v", actionID, err)
				}
			}
		}
	}

	// Persist task spec and create run model
	run, err := s.persistRunModel(ctx, runId, taskID, taskSpec, inputPrefix, runOutputBase, runSpec, request.GetSource(), triggerName, triggerTaskName, triggerRevision, triggerType)
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

// persistRunModel stores the task spec, builds the run model, and inserts it to DB via CreateAction.
// triggerName should be non-empty when the run originates from a scheduled trigger; it is stored on
// the run model and triggers a triggered_at update on the trigger row.
func (s *RunService) persistRunModel(
	ctx context.Context,
	runId *common.RunIdentifier,
	taskID *task.TaskIdentifier,
	taskSpec *task.TaskSpec,
	inputPrefix, runOutputBase string,
	runSpec *task.RunSpec,
	source workflow.RunSource,
	triggerName, triggerTaskName string,
	triggerRevision int64,
	triggerType string,
) (*models.Run, error) {
	// Store task spec and compute digest
	info := &workflow.RunInfo{InputsUri: inputPrefix}
	taskSpecModel, err := models.NewTaskSpecModel(ctx, taskSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to create task spec model: %w", err)
	}
	if taskSpecModel != nil {
		if err := s.repo.TaskRepo().CreateTaskSpec(ctx, taskSpecModel); err != nil {
			logger.Warnf(ctx, "CreateRun: failed to store task spec: %v", err)
		} else {
			info.TaskSpecDigest = taskSpecModel.Digest
		}
	}
	detailedInfo, err := proto.Marshal(info)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal run info: %w", err)
	}

	// Build ActionSpec
	actionSpec := &workflow.ActionSpec{
		ActionId:         &common.ActionIdentifier{Run: runId, Name: RootActionName},
		ParentActionName: nil,
		RunSpec:          runSpec,
		InputUri:         inputPrefix + "/inputs.pb",
		RunOutputBase:    runOutputBase,
		Spec: &workflow.ActionSpec_Task{
			Task: &workflow.TaskAction{
				Id:   taskID,
				Spec: taskSpec,
			},
		},
	}
	actionSpecBytes, err := proto.Marshal(actionSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal action spec: %w", err)
	}

	var runSpecBytes []byte
	if runSpec != nil {
		runSpecBytes, err = proto.Marshal(runSpec)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal run spec: %w", err)
		}
	}

	nullStr := func(s string) sql.NullString {
		return sql.NullString{String: s, Valid: s != ""}
	}

	runModel := &models.Run{
		Project:         runId.GetProject(),
		Domain:          runId.GetDomain(),
		RunName:         runId.GetName(),
		Name:            RootActionName,
		Phase:           int32(common.ActionPhase_ACTION_PHASE_QUEUED),
		ActionType:      int32(workflow.ActionType_ACTION_TYPE_TASK),
		TaskProject:     nullStr(taskID.GetProject()),
		TaskDomain:      nullStr(taskID.GetDomain()),
		TaskName:        nullStr(taskID.GetName()),
		TaskVersion:     nullStr(taskID.GetVersion()),
		TaskType:        taskSpec.GetTaskTemplate().GetType(),
		TaskShortName:   nullStr(taskSpec.GetShortName()),
		FunctionName:    taskID.GetName(),
		EnvironmentName: nullStr(taskSpec.GetEnvironment().GetName()),
		ActionSpec:      actionSpecBytes,
		ActionDetails:   []byte("{}"),
		DetailedInfo:    detailedInfo,
		RunSpec:         runSpecBytes,
		Attempts:        1,
		RunSource:       source.String(),
		TriggerTaskName: nullStr(triggerTaskName),
		TriggerName:     nullStr(triggerName),
		TriggerRevision: sql.NullInt64{Int64: triggerRevision, Valid: triggerRevision != 0},
		TriggerType:     nullStr(triggerType),
	}

	return s.repo.ActionRepo().CreateAction(ctx, runModel, triggerName != "")
}

// AbortRun aborts a run
func (s *RunService) AbortRun(
	ctx context.Context,
	req *connect.Request[workflow.AbortRunRequest],
) (*connect.Response[workflow.AbortRunResponse], error) {
	logger.Infof(ctx, "Received AbortRun request for run: %s/%s/%s",
		req.Msg.RunId.Project, req.Msg.RunId.Domain, req.Msg.RunId.Name)

	// Validate request
	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	reason := "User requested abort"
	if req.Msg.Reason != nil {
		reason = *req.Msg.Reason
	}

	// Mark only the root action ABORTED in DB, then push it to the reconciler.
	// The reconciler deletes "a0"'s CRD; K8s cascades deletion to all child CRDs
	// via OwnerReferences, and the action service informer marks them ABORTED in DB.
	if err := s.repo.ActionRepo().AbortRun(ctx, req.Msg.RunId, reason, nil); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("run not found: %s/%s/%s", req.Msg.RunId.Project, req.Msg.RunId.Domain, req.Msg.RunId.Name))
		}
		logger.Errorf(ctx, "Failed to abort run: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if s.abortReconciler != nil {
		s.abortReconciler.Push(ctx, &common.ActionIdentifier{Run: req.Msg.RunId, Name: "a0"}, reason)
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
		RunSpec: extractRunSpecFromActionModel(ctx, run),
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
		abortInfo := &workflow.AbortInfo{}
		if model.AbortReason != nil {
			abortInfo.Reason = *model.AbortReason
		}
		action.Result = &workflow.ActionDetails_AbortInfo{
			AbortInfo: abortInfo,
		}
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
	// Order events by reported time, falling back to updated time.
	sort.SliceStable(events, func(i, j int) bool {
		reportedTimeI := events[i].GetReportedTime()
		reportedTimeJ := events[j].GetReportedTime()
		if reportedTimeI != nil && reportedTimeJ != nil {
			return reportedTimeI.AsTime().Before(reportedTimeJ.AsTime())
		}
		return events[i].GetUpdatedTime().AsTime().Before(events[j].GetUpdatedTime().AsTime())
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

func IsTerminalPhase(phase common.ActionPhase) bool {
	return phase == common.ActionPhase_ACTION_PHASE_FAILED ||
		phase == common.ActionPhase_ACTION_PHASE_SUCCEEDED ||
		phase == common.ActionPhase_ACTION_PHASE_TIMED_OUT ||
		phase == common.ActionPhase_ACTION_PHASE_ABORTED
}

// lastAttemptIsTerminal returns true when the highest-numbered attempt has reached a
// terminal phase. Used by WatchActionDetails to close the stream only after action_events
// reflects the terminal transition, not just the actions table.
func lastAttemptIsTerminal(attempts []*workflow.ActionAttempt) bool {
	if len(attempts) == 0 {
		return false
	}
	var last *workflow.ActionAttempt
	for _, a := range attempts {
		if last == nil || a.GetAttempt() > last.GetAttempt() {
			last = a
		}
	}
	return IsTerminalPhase(last.GetPhase())
}

// GetActionData keeps backward compatibility by delegating data reads to DataProxy.
func (s *RunService) GetActionData(
	ctx context.Context,
	req *connect.Request[workflow.GetActionDataRequest],
) (*connect.Response[workflow.GetActionDataResponse], error) {
	logger.Infof(ctx, "Received GetActionData request for: %s/%s",
		req.Msg.ActionId.Run.Name, req.Msg.ActionId.Name)

	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if s.dataProxyClient == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("dataproxy client is not configured"))
	}

	dpResp, err := s.dataProxyClient.GetActionData(ctx, connect.NewRequest(&dataproxy.GetActionDataRequest{
		ActionId: req.Msg.GetActionId(),
	}))
	if err != nil {
		return nil, err
	}

	resp := &workflow.GetActionDataResponse{
		Inputs:  dpResp.Msg.GetInputs(),
		Outputs: dpResp.Msg.GetOutputs(),
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
	logger.Debugf(ctx, "Received ListRuns request")

	// Validate request
	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Build scope filter: always restrict to root actions (runs).
	scopeFilter := interfaces.Filter(impl.NewIsRootActionFilter())
	switch scope := req.Msg.ScopeBy.(type) {
	case *workflow.ListRunsRequest_ProjectId:
		scopeFilter = scopeFilter.And(impl.NewProjectIdFilter(scope.ProjectId))
	case *workflow.ListRunsRequest_TriggerName:
		scopeFilter = scopeFilter.And(impl.NewTriggerNameFilter(scope.TriggerName))
	case *workflow.ListRunsRequest_TaskName:
		scopeFilter = scopeFilter.And(impl.NewRunTaskNameFilter(scope.TaskName))
	case *workflow.ListRunsRequest_TaskId:
		scopeFilter = scopeFilter.And(impl.NewRunTaskIdFilter(scope.TaskId))
	}

	// Parse pagination, sort, and user-supplied filters from the common ListRequest.
	listInput, err := impl.NewListResourceInputFromProto(req.Msg.Request, models.ActionColumnsSet)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Merge: scope filter always takes precedence, user filters are ANDed on top.
	if listInput.Filter == nil {
		listInput.Filter = scopeFilter
	} else {
		listInput.Filter = scopeFilter.And(listInput.Filter)
	}

	actions, err := s.repo.ActionRepo().ListActions(ctx, listInput)
	if err != nil {
		logger.Errorf(ctx, "Failed to list runs: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// We fetch Limit+1 rows to detect whether a next page exists without a
	// separate COUNT query. If we got more than Limit rows back, there is at
	// least one more page: trim the slice and encode the last returned row's
	// created_at as the keyset cursor for the next request.
	var nextToken string
	if len(actions) > listInput.Limit {
		actions = actions[:listInput.Limit]
		nextToken = actions[len(actions)-1].CreatedAt.UTC().Format(time.RFC3339Nano)
	}

	protoRuns := make([]*workflow.Run, len(actions))
	for i, run := range actions {
		protoRuns[i] = s.convertRunToProto(run)
	}

	logger.Debugf(ctx, "Listed %d runs", len(actions))
	return connect.NewResponse(&workflow.ListRunsResponse{
		Runs:  protoRuns,
		Token: nextToken,
	}), nil
}

// ListActions lists actions for a run
func (s *RunService) ListActions(
	ctx context.Context,
	req *connect.Request[workflow.ListActionsRequest],
) (*connect.Response[workflow.ListActionsResponse], error) {
	logger.Infof(ctx, "Received ListActions request")

	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	listInput, err := impl.NewListResourceInputFromProto(req.Msg.Request, models.ActionColumnsSet)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	runFilter := interfaces.Filter(impl.NewRunActionsFilter(req.Msg.RunId))
	if listInput.Filter == nil {
		listInput.Filter = runFilter
	} else {
		listInput.Filter = runFilter.And(listInput.Filter)
	}

	actions, err := s.repo.ActionRepo().ListActions(ctx, listInput)
	if err != nil {
		logger.Errorf(ctx, "Failed to list actions: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// We fetch Limit+1 rows to detect whether a next page exists without a
	// separate COUNT query. If we got more than Limit rows back, there is at
	// least one more page: trim the slice and encode the last returned row's
	// created_at as the keyset cursor for the next request.
	var nextToken string
	if len(actions) > listInput.Limit {
		actions = actions[:listInput.Limit]
		nextToken = actions[len(actions)-1].CreatedAt.UTC().Format(time.RFC3339Nano)
	}

	protoActions := make([]*workflow.Action, len(actions))
	for i, action := range actions {
		ea := s.convertActionToEnrichedProto(action)
		protoActions[i] = ea.Action
	}

	logger.Infof(ctx, "Listed %d actions", len(actions))
	return connect.NewResponse(&workflow.ListActionsResponse{
		Actions: protoActions,
		Token:   nextToken,
	}), nil
}

func (s *RunService) GetActionDataURIs(
	ctx context.Context,
	req *connect.Request[workflow.GetActionDataURIsRequest],
) (*connect.Response[workflow.GetActionDataURIsResponse], error) {
	action, err := s.repo.ActionRepo().GetAction(ctx, req.Msg.GetActionId())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("action not found: %w", err))
	}

	if len(action.DetailedInfo) == 0 {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("run data not available"))
	}

	info := &workflow.RunInfo{}
	if err := proto.Unmarshal(action.DetailedInfo, info); err != nil {
		return nil, err
	}

	resp := &workflow.GetActionDataURIsResponse{
		InputsUri: info.GetInputsUri(),
	}

	if action.Phase == int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED) {
		if workflow.ActionType(action.ActionType) == workflow.ActionType_ACTION_TYPE_TRACE {
			resp.OutputsUri = info.GetOutputsUri()
		} else {
			attempts, err := s.getAttempts(ctx, req.Msg.GetActionId())
			if err != nil {
				return nil, err
			}
			if len(attempts) == 0 {
				return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("outputs not available, no attempts for action"))
			}
			outputUri := attempts[len(attempts)-1].GetOutputs().GetOutputUri()
			if outputUri == "" {
				return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("outputs not available"))
			}
			resp.OutputsUri = outputUri
		}
	}

	return connect.NewResponse(resp), nil
}

func (s *RunService) GetActionLogContext(
	ctx context.Context,
	req *connect.Request[workflow.GetActionLogContextRequest],
) (*connect.Response[workflow.GetActionLogContextResponse], error) {
	logContext, cluster, err := getLogContextAndClusterForAttempt(ctx, s.repo, req.Msg.GetActionId(), req.Msg.GetAttempt())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&workflow.GetActionLogContextResponse{
		LogContext: logContext,
		Cluster:    cluster,
	}), nil
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

	// Abort in database, then push to reconciler for background pod termination.
	if err := s.repo.ActionRepo().AbortAction(ctx, req.Msg.ActionId, reason, nil); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("action not found: %s", req.Msg.ActionId.Name))
		}
		logger.Errorf(ctx, "Failed to abort action: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if s.abortReconciler != nil {
		s.abortReconciler.Push(ctx, req.Msg.ActionId, reason)
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

	// Subscribe FIRST to avoid missing notifications that fire between the DB
	// read and the subscription setup (same pattern as WatchActions).
	updates := make(chan *models.Action, 50)
	errs := make(chan error, 1)
	go s.repo.ActionRepo().WatchActionUpdates(ctx, actionID, updates, errs)

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

	// Close only once action_events reflects the terminal phase, not just actions table.
	if lastAttemptIsTerminal(details.GetAttempts()) {
		return nil
	}

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
			// Close once action_events reflects the terminal phase.
			if lastAttemptIsTerminal(details.GetAttempts()) {
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

	// Step 1: Subscribe to updates first to avoid missing events that arrive
	// between the initial list and the subscription (same pattern as WatchActions).
	updatesCh := make(chan *models.Run, 50)
	errsCh := make(chan error, 1)
	go s.repo.ActionRepo().WatchAllRunUpdates(ctx, updatesCh, errsCh)

	// Step 2: Send existing runs that match filter
	listInput := s.convertWatchRequestToListInput(req.Msg)

	runs, err := s.repo.ActionRepo().ListActions(ctx, listInput)
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
	go s.repo.ActionRepo().WatchAllActionUpdates(ctx, runID, updatesCh, errsCh)

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
	const pageSize = 100
	offset := 0
	for {
		// Sort ascending by created_at so parent actions are inserted into the
		// run state manager before their children. insertAction requires the
		// parent node to already exist in the tree when a child is processed.
		batch, err := s.repo.ActionRepo().ListActions(ctx, interfaces.ListResourceInput{
			Filter: impl.NewRunActionsFilter(runID),
			Limit:  pageSize,
			Offset: offset,
			SortParameters: []interfaces.SortParameter{
				impl.NewSortParameter("created_at", interfaces.SortOrderAscending),
			},
		})
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

		if len(batch) < pageSize {
			return nil
		}
		offset += pageSize
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

// WatchClusterEvents streams Kubernetes cluster events recorded in action_events.
func (s *RunService) WatchClusterEvents(
	ctx context.Context,
	req *connect.Request[workflow.WatchClusterEventsRequest],
	stream *connect.ServerStream[workflow.WatchClusterEventsResponse],
) error {
	actionID := req.Msg.Id
	attempt := req.Msg.GetAttempt()
	logger.Infof(ctx, "Received WatchClusterEvents request for: %s/%s", actionID.Run.Name, actionID.Name)

	// Start watching first to reduce the chance of missing action updates
	// between initial state fetch and subscription setup.
	updatesCh := make(chan *models.Action, 50)
	errsCh := make(chan error, 1)
	go s.repo.ActionRepo().WatchActionUpdates(ctx, actionID, updatesCh, errsCh)

	action, err := s.repo.ActionRepo().GetAction(ctx, actionID)
	if err != nil {
		return connect.NewError(connect.CodeNotFound, fmt.Errorf("action not found: %w", err))
	}

	const maxEvents = 500
	var lastUpdatedAt time.Time

	for {
		offset := 0
		isTerminal := IsTerminalPhase(common.ActionPhase(action.Phase))

		// Drain all available cluster events for current action since lastUpdatedAt
		for {
			info, err := s.getClusterEventsInfo(ctx, actionID, attempt, lastUpdatedAt, offset, maxEvents)
			if err != nil {
				return connect.NewError(connect.CodeInternal, err)
			}
			isTerminal = isTerminal || info.isTerminal

			if len(info.events) > 0 {
				if err := stream.Send(&workflow.WatchClusterEventsResponse{ClusterEvents: info.events}); err != nil {
					return err
				}
			}

			if info.lastUpdatedAt.After(lastUpdatedAt) {
				lastUpdatedAt = info.lastUpdatedAt
			}

			if len(info.events) < maxEvents {
				break
			}
			offset += maxEvents
		}
		if isTerminal {
			return nil
		}

		select {
		case <-ctx.Done():
			return nil
		case err := <-errsCh:
			return connect.NewError(connect.CodeInternal, err)
		case updated, ok := <-updatesCh:
			if !ok {
				return nil
			}
			action = updated
		}
	}
}

type clusterEventsInfo struct {
	events        []*workflow.ClusterEvent
	lastUpdatedAt time.Time
	isTerminal    bool
}

func (s *RunService) getClusterEventsInfo(
	ctx context.Context,
	actionID *common.ActionIdentifier,
	attempt uint32,
	since time.Time,
	offset, limit int,
) (clusterEventsInfo, error) {
	eventModels, err := s.repo.ActionRepo().ListEventsSince(ctx, actionID, attempt, since, offset, limit)
	if err != nil {
		return clusterEventsInfo{}, fmt.Errorf("failed to list action events since %s: %w", since.Format(time.RFC3339Nano), err)
	}

	var info clusterEventsInfo
	for _, model := range eventModels {
		event, err := model.ToActionEvent()
		if err != nil {
			return clusterEventsInfo{}, fmt.Errorf("failed to convert action event: %w", err)
		}
		if model.UpdatedAt.After(info.lastUpdatedAt) {
			info.lastUpdatedAt = model.UpdatedAt
		}
		info.events = append(info.events, event.GetClusterEvents()...)
		if IsTerminalPhase(event.GetPhase()) {
			info.isTerminal = true
		}
	}

	return info, nil
}

// actionModelToDetails converts a DB Action model to an ActionDetails proto.
func (s *RunService) actionModelToDetails(action *models.Action, actionID *common.ActionIdentifier) *workflow.ActionDetails {
	status := &workflow.ActionStatus{
		Phase:       common.ActionPhase(action.Phase),
		StartTime:   timestamppb.New(action.CreatedAt),
		Attempts:    action.Attempts,
		CacheStatus: action.CacheStatus,
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

// getLogContextAndClusterForAttempt is like getLogContextForAttempt but also returns the cluster name.
func getLogContextAndClusterForAttempt(ctx context.Context, repo interfaces.Repository, actionID *common.ActionIdentifier, attempt uint32) (*core.LogContext, string, error) {
	m, err := repo.ActionRepo().GetLatestEventByAttempt(ctx, actionID, attempt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", connect.NewError(connect.CodeNotFound, fmt.Errorf("no event found for action %v attempt %d", actionID, attempt))
		}
		return nil, "", connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get event for action %v attempt %d: %w", actionID, attempt, err))
	}

	event, err := m.ToActionEvent()
	if err != nil {
		return nil, "", connect.NewError(connect.CodeInternal, fmt.Errorf("failed to deserialize event: %w", err))
	}

	if event.GetLogContext() == nil {
		return nil, "", connect.NewError(connect.CodeNotFound, fmt.Errorf("no log context found for action %v attempt %d", actionID, attempt))
	}

	return event.GetLogContext(), event.GetCluster(), nil
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

	if action.RunSource != "" {
		if v, ok := workflow.RunSource_value[action.RunSource]; ok {
			metadata.Source = workflow.RunSource(v)
		}
	}

	if action.TriggerName.Valid {
		metadata.TriggerName = action.TriggerName.String
		metadata.TriggerId = &common.TriggerIdentifier{
			Name: &common.TriggerName{
				Project:  action.Project,
				Domain:   action.Domain,
				Name:     action.TriggerName.String,
				TaskName: action.TriggerTaskName.String,
			},
		}
		if action.TriggerRevision.Valid {
			metadata.TriggerId.Revision = uint64(action.TriggerRevision.Int64)
		}
	}

	switch workflow.ActionType(action.ActionType) {
	case workflow.ActionType_ACTION_TYPE_TASK:
		taskMeta := &workflow.TaskActionMetadata{
			TaskType: action.TaskType,
		}
		if action.TaskName.Valid {
			taskMeta.Id = &task.TaskIdentifier{
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
	if err := proto.Unmarshal(specProto, &specMsg); err == nil {
		return &specMsg
	}
	return nil
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

func extractRunSpecFromActionModel(ctx context.Context, action *models.Action) *task.RunSpec {
	if action == nil {
		return nil
	}

	if len(action.RunSpec) > 0 {
		spec := &task.RunSpec{}
		if err := proto.Unmarshal(action.RunSpec, spec); err == nil {
			return spec
		} else {
			logger.Warnf(ctx, "Failed to unmarshal action.RunSpec for %s/%s/%s/%s, falling back to ActionSpec: %v",
				action.Project, action.Domain, action.RunName, action.Name, err)
		}
	}

	// Fallback to extract from ActionSpec
	return extractRunSpec(action.ActionSpec)
}

// Helper functions

// buildInputPrefix generates the input path prefix for the root action.
func buildInputPrefix(storagePrefix, project, domain, name string) string {
	return fmt.Sprintf("%s/%s/%s/%s/inputs",
		strings.TrimRight(storagePrefix, "/"),
		project, domain, name)
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
func buildRunOutputBase(storagePrefix, project, domain, name string) string {
	return fmt.Sprintf("%s/%s/%s/%s/",
		strings.TrimRight(storagePrefix, "/"),
		project, domain, name)
}

// convertRunToProto converts a repository Run to a proto Run
func (s *RunService) convertRunToProto(run *models.Run) *workflow.Run {
	if run == nil {
		return nil
	}

	runID := &common.RunIdentifier{
		Project: run.Project,
		Domain:  run.Domain,
		Name:    run.RunName,
	}

	action := &workflow.Action{
		Id: &common.ActionIdentifier{
			Run:  runID,
			Name: run.Name,
		},
		Metadata: actionMetadataFromModel(run),
		Status: &workflow.ActionStatus{
			Phase:       common.ActionPhase(run.Phase),
			StartTime:   timestamppb.New(run.CreatedAt),
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
			Project: action.Project,
			Domain:  action.Domain,
			Name:    action.RunName,
		},
		Name: action.Name,
	}

	actionStatus := &workflow.ActionStatus{
		Phase:       common.ActionPhase(action.Phase),
		Attempts:    action.Attempts,
		StartTime:   timestamppb.New(action.CreatedAt),
		CacheStatus: action.CacheStatus,
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
			Project: runID.Project,
			Domain:  runID.Domain,
			Name:    runID.Name,
		},
		Name: action.Name,
	}

	actionStatus := &workflow.ActionStatus{
		Phase:       common.ActionPhase(action.Phase),
		StartTime:   timestamppb.New(action.CreatedAt),
		Attempts:    action.Attempts,
		CacheStatus: action.CacheStatus,
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
	metadata.Source = workflow.RunSource(workflow.RunSource_value[action.RunSource])

	if action.TriggerName.Valid {
		metadata.TriggerName = action.TriggerName.String
		metadata.TriggerId = &common.TriggerIdentifier{
			Name: &common.TriggerName{
				Project:  action.Project,
				Domain:   action.Domain,
				Name:     action.TriggerName.String,
				TaskName: action.TriggerTaskName.String,
			},
		}
		if action.TriggerRevision.Valid {
			metadata.TriggerId.Revision = uint64(action.TriggerRevision.Int64)
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
	var project, domain string
	if pid := req.GetProjectId(); pid != nil {
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

	rootActions, err := s.repo.ActionRepo().ListRootActions(ctx, project, domain, startDate, endDate, 1000)
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
		Project: taskSpec.GetTaskTemplate().GetId().GetProject(),
		Domain:  taskSpec.GetTaskTemplate().GetId().GetDomain(),
		Name:    taskSpec.GetTaskTemplate().GetId().GetName(),
		Version: taskSpec.GetTaskTemplate().GetId().GetVersion(),
	}
}

// convertWatchRequestToListInput converts a WatchRunsRequest to a ListResourceInput for the initial run snapshot.
func (s *RunService) convertWatchRequestToListInput(req *workflow.WatchRunsRequest) interfaces.ListResourceInput {
	scopeFilter := interfaces.Filter(impl.NewIsRootActionFilter())
	switch target := req.Target.(type) {
	case *workflow.WatchRunsRequest_ProjectId:
		scopeFilter = scopeFilter.And(impl.NewProjectIdFilter(target.ProjectId))
	}
	return interfaces.ListResourceInput{
		Limit:  100,
		Filter: scopeFilter,
	}
}

// runMatchesFilter checks if a run matches the WatchRunsRequest filter criteria
func (s *RunService) runMatchesFilter(run *models.Run, req *workflow.WatchRunsRequest) bool {
	if req.Target == nil {
		return true
	}

	switch target := req.Target.(type) {
	case *workflow.WatchRunsRequest_ProjectId:
		return run.Project == target.ProjectId.Name &&
			run.Domain == target.ProjectId.Domain
	default:
		return true
	}
}
