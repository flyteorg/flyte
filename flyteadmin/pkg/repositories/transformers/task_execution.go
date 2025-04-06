package transformers

import (
	"context"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"sort"
	"strconv"

	jsonpatch "github.com/evanphx/json-patch"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	_struct "google.golang.org/protobuf/types/known/structpb"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

var empty _struct.Struct
var jsonEmpty, _ = protojson.Marshal(&empty)

type CreateTaskExecutionModelInput struct {
	Request               *admin.TaskExecutionEventRequest
	InlineEventDataPolicy interfaces.InlineEventDataPolicy
	StorageClient         *storage.DataStore
}

func addTaskStartedState(request *admin.TaskExecutionEventRequest, taskExecutionModel *models.TaskExecution,
	closure *admin.TaskExecutionClosure) error {
	occurredAt := request.GetEvent().GetOccurredAt().AsTime()
	//Updated the startedAt timestamp only if its not set.
	// The task start event should already be updating this through addTaskStartedState
	// This check makes sure any out of order
	if taskExecutionModel.StartedAt == nil {
		taskExecutionModel.StartedAt = &occurredAt
		closure.StartedAt = request.GetEvent().GetOccurredAt()
	}
	return nil
}

func addTaskTerminalState(
	ctx context.Context,
	request *admin.TaskExecutionEventRequest,
	taskExecutionModel *models.TaskExecution, closure *admin.TaskExecutionClosure,
	inlineEventDataPolicy interfaces.InlineEventDataPolicy, storageClient *storage.DataStore) error {
	if taskExecutionModel.StartedAt == nil {
		logger.Warning(context.Background(), "task execution is missing StartedAt")
	} else {
		endTime := request.GetEvent().GetOccurredAt().AsTime()
		closure.StartedAt = timestamppb.New(*taskExecutionModel.StartedAt)
		taskExecutionModel.Duration = endTime.Sub(*taskExecutionModel.StartedAt)
		closure.Duration = durationpb.New(taskExecutionModel.Duration)
	}

	if request.GetEvent().GetOutputUri() != "" {
		closure.OutputResult = &admin.TaskExecutionClosure_OutputUri{
			OutputUri: request.GetEvent().GetOutputUri(),
		}
	} else if request.GetEvent().GetOutputData() != nil {
		switch inlineEventDataPolicy {
		case interfaces.InlineEventDataPolicyStoreInline:
			closure.OutputResult = &admin.TaskExecutionClosure_OutputData{
				OutputData: request.GetEvent().GetOutputData(),
			}
		default:
			logger.Debugf(ctx, "Offloading outputs per InlineEventDataPolicy")
			uri, err := common.OffloadLiteralMap(ctx, storageClient, request.GetEvent().GetOutputData(),
				request.GetEvent().GetParentNodeExecutionId().GetExecutionId().GetProject(), request.GetEvent().GetParentNodeExecutionId().GetExecutionId().GetDomain(),
				request.GetEvent().GetParentNodeExecutionId().GetExecutionId().GetName(), request.GetEvent().GetParentNodeExecutionId().GetNodeId(),
				request.GetEvent().GetTaskId().GetProject(), request.GetEvent().GetTaskId().GetDomain(), request.GetEvent().GetTaskId().GetName(), request.GetEvent().GetTaskId().GetVersion(),
				strconv.FormatUint(uint64(request.GetEvent().GetRetryAttempt()), 10), OutputsObjectSuffix)
			if err != nil {
				return err
			}
			closure.OutputResult = &admin.TaskExecutionClosure_OutputUri{
				OutputUri: uri.String(),
			}
		}
	} else if request.GetEvent().GetError() != nil {
		closure.OutputResult = &admin.TaskExecutionClosure_Error{
			Error: request.GetEvent().GetError(),
		}
	}
	return nil
}

func CreateTaskExecutionModel(ctx context.Context, input CreateTaskExecutionModelInput) (*models.TaskExecution, error) {
	taskExecution := &models.TaskExecution{
		TaskExecutionKey: models.TaskExecutionKey{
			TaskKey: models.TaskKey{
				Project: input.Request.GetEvent().GetTaskId().GetProject(),
				Domain:  input.Request.GetEvent().GetTaskId().GetDomain(),
				Name:    input.Request.GetEvent().GetTaskId().GetName(),
				Version: input.Request.GetEvent().GetTaskId().GetVersion(),
			},
			NodeExecutionKey: models.NodeExecutionKey{
				NodeID: input.Request.GetEvent().GetParentNodeExecutionId().GetNodeId(),
				ExecutionKey: models.ExecutionKey{
					Project: input.Request.GetEvent().GetParentNodeExecutionId().GetExecutionId().GetProject(),
					Domain:  input.Request.GetEvent().GetParentNodeExecutionId().GetExecutionId().GetDomain(),
					Name:    input.Request.GetEvent().GetParentNodeExecutionId().GetExecutionId().GetName(),
				},
			},
			RetryAttempt: &input.Request.Event.RetryAttempt,
		},

		Phase:        input.Request.GetEvent().GetPhase().String(),
		PhaseVersion: input.Request.GetEvent().GetPhaseVersion(),
	}
	err := handleTaskExecutionInputs(ctx, taskExecution, input.Request, input.StorageClient)
	if err != nil {
		return nil, err
	}

	metadata := input.Request.GetEvent().GetMetadata()
	if metadata != nil && len(metadata.GetExternalResources()) > 1 {
		sort.Slice(metadata.GetExternalResources(), func(i, j int) bool {
			a := metadata.GetExternalResources()[i]
			b := metadata.GetExternalResources()[j]
			if a.GetIndex() == b.GetIndex() {
				return a.GetRetryAttempt() < b.GetRetryAttempt()
			}
			return a.GetIndex() < b.GetIndex()
		})
	}

	reportedAt := input.Request.GetEvent().GetReportedAt()
	if reportedAt == nil || (reportedAt.GetSeconds() == 0 && reportedAt.GetNanos() == 0) {
		reportedAt = input.Request.GetEvent().GetOccurredAt()
	}

	closure := &admin.TaskExecutionClosure{
		Phase:        input.Request.GetEvent().GetPhase(),
		UpdatedAt:    reportedAt,
		CreatedAt:    input.Request.GetEvent().GetOccurredAt(),
		Logs:         input.Request.GetEvent().GetLogs(),
		CustomInfo:   input.Request.GetEvent().GetCustomInfo(),
		TaskType:     input.Request.GetEvent().GetTaskType(),
		Metadata:     metadata,
		EventVersion: input.Request.GetEvent().GetEventVersion(),
	}

	if len(input.Request.GetEvent().GetReasons()) > 0 {
		for _, reason := range input.Request.GetEvent().GetReasons() {
			closure.Reasons = append(closure.GetReasons(), &admin.Reason{
				OccurredAt: reason.GetOccurredAt(),
				Message:    reason.GetReason(),
			})
		}
		closure.Reason = input.Request.GetEvent().GetReasons()[len(input.Request.GetEvent().GetReasons())-1].GetReason()
	} else if len(input.Request.GetEvent().GetReason()) > 0 {
		closure.Reasons = []*admin.Reason{
			{
				OccurredAt: input.Request.GetEvent().GetOccurredAt(),
				Message:    input.Request.GetEvent().GetReason(),
			},
		}
		closure.Reason = input.Request.GetEvent().GetReason()
	}

	eventPhase := input.Request.GetEvent().GetPhase()

	// Different tasks may report different phases as their first event.
	// If the first event we receive for this execution is a valid
	// non-terminal phase, mark the execution start time.
	if eventPhase == core.TaskExecution_RUNNING {
		err := addTaskStartedState(input.Request, taskExecution, closure)
		if err != nil {
			return nil, err
		}
	}

	if common.IsTaskExecutionTerminal(input.Request.GetEvent().GetPhase()) {
		err := addTaskTerminalState(ctx, input.Request, taskExecution, closure, input.InlineEventDataPolicy, input.StorageClient)
		if err != nil {
			return nil, err
		}
	}
	marshaledClosure, err := proto.Marshal(closure)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(
			codes.Internal, "failed to marshal task execution closure with error: %v", err)
	}

	taskExecution.Closure = marshaledClosure
	taskExecutionCreatedAt := input.Request.GetEvent().GetOccurredAt().AsTime()
	taskExecution.TaskExecutionCreatedAt = &taskExecutionCreatedAt
	taskExecutionUpdatedAt := reportedAt.AsTime()
	taskExecution.TaskExecutionUpdatedAt = &taskExecutionUpdatedAt

	return taskExecution, nil
}

// mergeLogs returns the unique list of logs across an existing list and the latest list sent in a task execution event
// update.
// It returns all the new logs receives + any existing log that hasn't been overwritten by a new log.
// An existing logLink is said to have been overwritten if a new logLink with the same Uri or the same Name has been
// received.
func mergeLogs(existing, latest []*core.TaskLog) []*core.TaskLog {
	if len(latest) == 0 {
		return existing
	}

	if len(existing) == 0 {
		return latest
	}

	latestSetByURI := make(map[string]*core.TaskLog, len(latest))
	latestSetByName := make(map[string]*core.TaskLog, len(latest))
	for _, latestLog := range latest {
		latestSetByURI[latestLog.GetUri()] = latestLog
		if len(latestLog.GetName()) > 0 {
			latestSetByName[latestLog.GetName()] = latestLog
		}
	}

	// Copy over the latest logs since names will change for existing logs as a task transitions across phases.
	logs := latest
	for _, existingLog := range existing {
		if _, ok := latestSetByURI[existingLog.GetUri()]; !ok {
			if _, ok = latestSetByName[existingLog.GetName()]; !ok {
				// We haven't seen this log before: add it to the output result list.
				logs = append(logs, existingLog)
			}
		}
	}

	return logs
}

func mergeCustom(existing, latest *_struct.Struct) (*_struct.Struct, error) {
	if existing == nil {
		return latest, nil
	}
	if latest == nil {
		return existing, nil
	}

	// To merge latest into existing we first create a patch object that consists of applying changes from latest to
	// an empty struct. Then we apply this patch to existing so that the values changed in latest take precedence but
	// barring conflicts/overwrites the values in existing stay the same.
	jsonExisting, err := protojson.Marshal(existing)
	if err != nil {
		return nil, err
	}
	jsonLatest, err := protojson.Marshal(latest)
	if err != nil {
		return nil, err
	}
	patch, err := jsonpatch.CreateMergePatch(jsonEmpty, jsonLatest)
	if err != nil {
		return nil, err
	}
	custom, err := jsonpatch.MergePatch(jsonExisting, patch)
	if err != nil {
		return nil, err
	}
	var response _struct.Struct

	err = protojson.Unmarshal(custom, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// mergeExternalResource combines the latest ExternalResourceInfo proto with an existing instance
// by updating fields and merging logs.
func mergeExternalResource(existing, latest *event.ExternalResourceInfo) *event.ExternalResourceInfo {
	if existing == nil {
		return latest
	}

	if latest == nil {
		return existing
	}

	if latest.GetExternalId() != "" && existing.GetExternalId() != latest.GetExternalId() {
		existing.ExternalId = latest.GetExternalId()
	}
	// note we are not updating existing.Index and existing.RetryAttempt because they are the
	// search key for our ExternalResource pool.
	existing.Phase = latest.GetPhase()
	if latest.GetCacheStatus() != core.CatalogCacheStatus_CACHE_DISABLED && existing.GetCacheStatus() != latest.GetCacheStatus() {
		existing.CacheStatus = latest.GetCacheStatus()
	}
	existing.Logs = mergeLogs(existing.GetLogs(), latest.GetLogs())

	// Overwrite custom info if provided
	if latest.GetCustomInfo() != nil {
		existing.CustomInfo = proto.Clone(latest.GetCustomInfo()).(*structpb.Struct)
	}

	return existing
}

// mergeExternalResources combines lists of external resources. This involves appending new
// resources and updating in-place resources attributes.
func mergeExternalResources(existing, latest []*event.ExternalResourceInfo) []*event.ExternalResourceInfo {
	if len(latest) == 0 {
		return existing
	}

	for _, externalResource := range latest {
		// we use a binary search over the ExternalResource Index and RetryAttempt fields to
		// determine if a new subtask is being reported or an existing is being updated. it is
		// important to note that this means anytime more than one ExternalResource is reported
		// they must set the Index field.
		index := sort.Search(len(existing), func(i int) bool {
			if existing[i].GetIndex() == externalResource.GetIndex() {
				return existing[i].GetRetryAttempt() >= externalResource.GetRetryAttempt()
			}
			return existing[i].GetIndex() >= externalResource.GetIndex()
		})

		if index >= len(existing) {
			existing = append(existing, externalResource)
		} else if existing[index].GetIndex() == externalResource.GetIndex() && existing[index].GetRetryAttempt() == externalResource.GetRetryAttempt() {
			existing[index] = mergeExternalResource(existing[index], externalResource)
		} else {
			existing = append(existing, &event.ExternalResourceInfo{})
			copy(existing[index+1:], existing[index:])
			existing[index] = externalResource
		}
	}

	return existing
}

// mergeMetadata merges an existing TaskExecutionMetadata instance with the provided instance. This
// includes updating non-defaulted fields and merging ExternalResources.
func mergeMetadata(existing, latest *event.TaskExecutionMetadata) *event.TaskExecutionMetadata {
	if existing == nil {
		return latest
	}

	if latest == nil {
		return existing
	}

	if latest.GetGeneratedName() != "" && existing.GetGeneratedName() != latest.GetGeneratedName() {
		existing.GeneratedName = latest.GetGeneratedName()
	}
	existing.ExternalResources = mergeExternalResources(existing.GetExternalResources(), latest.GetExternalResources())
	existing.ResourcePoolInfo = latest.GetResourcePoolInfo()
	if latest.GetPluginIdentifier() != "" && existing.GetPluginIdentifier() != latest.GetPluginIdentifier() {
		existing.PluginIdentifier = latest.GetPluginIdentifier()
	}
	if latest.GetInstanceClass() != event.TaskExecutionMetadata_DEFAULT && existing.GetInstanceClass() != latest.GetInstanceClass() {
		existing.InstanceClass = latest.GetInstanceClass()
	}

	return existing
}

func filterExternalResourceLogsByPhase(externalResources []*event.ExternalResourceInfo, phase core.TaskExecution_Phase) {
	for _, externalResource := range externalResources {
		externalResource.Logs = filterLogsByPhase(externalResource.GetLogs(), phase)
	}
}

func filterLogsByPhase(logs []*core.TaskLog, phase core.TaskExecution_Phase) []*core.TaskLog {
	filteredLogs := make([]*core.TaskLog, 0, len(logs))

	for _, l := range logs {
		if common.IsTaskExecutionTerminal(phase) && l.GetHideOnceFinished() {
			continue
		}
		// Some plugins like e.g. Dask, Ray start with or very quickly transition to core.TaskExecution_INITIALIZING
		// once the CR has been created even though the underlying pods are still pending. We thus treat queued and
		// initializing the same here.
		if (phase == core.TaskExecution_QUEUED || phase == core.TaskExecution_INITIALIZING) && !l.GetShowWhilePending() {
			continue
		}
		filteredLogs = append(filteredLogs, l)

	}
	return filteredLogs
}

func UpdateTaskExecutionModel(ctx context.Context, request *admin.TaskExecutionEventRequest, taskExecutionModel *models.TaskExecution,
	inlineEventDataPolicy interfaces.InlineEventDataPolicy, storageClient *storage.DataStore) error {
	err := handleTaskExecutionInputs(ctx, taskExecutionModel, request, storageClient)
	if err != nil {
		return err
	}
	var taskExecutionClosure admin.TaskExecutionClosure
	err = proto.Unmarshal(taskExecutionModel.Closure, &taskExecutionClosure)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal,
			"failed to unmarshal task execution closure with error: %+v", err)
	}
	isPhaseChange := taskExecutionModel.Phase != request.GetEvent().GetPhase().String()
	existingTaskPhase := taskExecutionModel.Phase
	taskExecutionModel.Phase = request.GetEvent().GetPhase().String()
	taskExecutionModel.PhaseVersion = request.GetEvent().GetPhaseVersion()
	taskExecutionClosure.Phase = request.GetEvent().GetPhase()
	reportedAt := request.GetEvent().GetReportedAt()
	if reportedAt == nil || (reportedAt.GetSeconds() == 0 && reportedAt.GetNanos() == 0) {
		reportedAt = request.GetEvent().GetOccurredAt()
	}
	taskExecutionClosure.UpdatedAt = reportedAt

	mergedLogs := mergeLogs(taskExecutionClosure.GetLogs(), request.GetEvent().GetLogs())
	filteredLogs := filterLogsByPhase(mergedLogs, request.GetEvent().GetPhase())
	taskExecutionClosure.Logs = filteredLogs

	if len(request.GetEvent().GetReasons()) > 0 {
		for _, reason := range request.GetEvent().GetReasons() {
			taskExecutionClosure.Reasons = append(
				taskExecutionClosure.GetReasons(),
				&admin.Reason{
					OccurredAt: reason.GetOccurredAt(),
					Message:    reason.GetReason(),
				})
		}
		taskExecutionClosure.Reason = request.GetEvent().GetReasons()[len(request.GetEvent().GetReasons())-1].GetReason()
	} else if len(request.GetEvent().GetReason()) > 0 {
		if taskExecutionClosure.GetReason() != request.GetEvent().GetReason() {
			// by tracking a time-series of reasons we increase the size of the TaskExecutionClosure in scenarios where
			// a task reports a large number of unique reasons. if this size increase becomes problematic we this logic
			// will need to be revisited.
			taskExecutionClosure.Reasons = append(
				taskExecutionClosure.GetReasons(),
				&admin.Reason{
					OccurredAt: request.GetEvent().GetOccurredAt(),
					Message:    request.GetEvent().GetReason(),
				})
		}

		taskExecutionClosure.Reason = request.GetEvent().GetReason()
	}
	if existingTaskPhase != core.TaskExecution_RUNNING.String() && taskExecutionModel.Phase == core.TaskExecution_RUNNING.String() {
		err = addTaskStartedState(request, taskExecutionModel, &taskExecutionClosure)
		if err != nil {
			return err
		}
	}

	if common.IsTaskExecutionTerminal(request.GetEvent().GetPhase()) {
		err := addTaskTerminalState(ctx, request, taskExecutionModel, &taskExecutionClosure, inlineEventDataPolicy, storageClient)
		if err != nil {
			return err
		}
	}
	taskExecutionClosure.CustomInfo, err = mergeCustom(taskExecutionClosure.GetCustomInfo(), request.GetEvent().GetCustomInfo())
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal, "failed to merge task event custom_info with error: %v", err)
	}
	taskExecutionClosure.Metadata = mergeMetadata(taskExecutionClosure.GetMetadata(), request.GetEvent().GetMetadata())

	if isPhaseChange && taskExecutionClosure.GetMetadata() != nil && len(taskExecutionClosure.GetMetadata().GetExternalResources()) > 0 {
		filterExternalResourceLogsByPhase(taskExecutionClosure.GetMetadata().GetExternalResources(), request.GetEvent().GetPhase())
	}

	if request.GetEvent().GetEventVersion() > taskExecutionClosure.GetEventVersion() {
		taskExecutionClosure.EventVersion = request.GetEvent().GetEventVersion()
	}
	marshaledClosure, err := proto.Marshal(&taskExecutionClosure)
	if err != nil {
		return errors.NewFlyteAdminErrorf(
			codes.Internal, "failed to marshal task execution closure with error: %v", err)
	}
	taskExecutionModel.Closure = marshaledClosure
	updatedAt := reportedAt.AsTime()
	taskExecutionModel.TaskExecutionUpdatedAt = &updatedAt
	return nil
}

func FromTaskExecutionModel(taskExecutionModel models.TaskExecution, opts *ExecutionTransformerOptions) (*admin.TaskExecution, error) {
	var closure admin.TaskExecutionClosure
	err := proto.Unmarshal(taskExecutionModel.Closure, &closure)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to unmarshal closure")
	}
	if closure.GetError() != nil && opts != nil && opts.TrimErrorMessage && len(closure.GetError().GetMessage()) > 0 {
		trimmedErrOutputResult := closure.GetError()
		trimmedErrMessage := TrimErrorMessage(trimmedErrOutputResult.GetMessage())
		trimmedErrOutputResult.Message = trimmedErrMessage
		closure.OutputResult = &admin.TaskExecutionClosure_Error{
			Error: trimmedErrOutputResult,
		}
	}

	taskExecution := &admin.TaskExecution{
		Id: &core.TaskExecutionIdentifier{
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      taskExecutionModel.TaskExecutionKey.TaskKey.Project,
				Domain:       taskExecutionModel.TaskExecutionKey.TaskKey.Domain,
				Name:         taskExecutionModel.TaskExecutionKey.TaskKey.Name,
				Version:      taskExecutionModel.TaskExecutionKey.TaskKey.Version,
			},
			NodeExecutionId: &core.NodeExecutionIdentifier{
				NodeId: taskExecutionModel.NodeExecutionKey.NodeID,
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: taskExecutionModel.TaskExecutionKey.NodeExecutionKey.ExecutionKey.Project,
					Domain:  taskExecutionModel.TaskExecutionKey.NodeExecutionKey.ExecutionKey.Domain,
					Name:    taskExecutionModel.TaskExecutionKey.NodeExecutionKey.ExecutionKey.Name,
				},
			},
			RetryAttempt: *taskExecutionModel.TaskExecutionKey.RetryAttempt,
		},
		InputUri: taskExecutionModel.InputURI,
		Closure:  &closure,
	}
	if len(taskExecutionModel.ChildNodeExecution) > 0 {
		taskExecution.IsParent = true
	}

	return taskExecution, nil
}

func FromTaskExecutionModels(taskExecutionModels []models.TaskExecution, opts *ExecutionTransformerOptions) ([]*admin.TaskExecution, error) {
	taskExecutions := make([]*admin.TaskExecution, len(taskExecutionModels))
	for idx, taskExecutionModel := range taskExecutionModels {
		taskExecution, err := FromTaskExecutionModel(taskExecutionModel, opts)
		if err != nil {
			return nil, err
		}
		taskExecutions[idx] = taskExecution
	}
	return taskExecutions, nil
}

func handleTaskExecutionInputs(ctx context.Context, taskExecutionModel *models.TaskExecution,
	request *admin.TaskExecutionEventRequest, storageClient *storage.DataStore) error {
	if len(taskExecutionModel.InputURI) > 0 {
		// Inputs are static over the duration of the task execution, no need to update them when they're already set
		return nil
	}
	switch request.GetEvent().GetInputValue().(type) {
	case *event.TaskExecutionEvent_InputUri:
		taskExecutionModel.InputURI = request.GetEvent().GetInputUri()
	case *event.TaskExecutionEvent_InputData:
		uri, err := common.OffloadLiteralMap(ctx, storageClient, request.GetEvent().GetInputData(),
			request.GetEvent().GetParentNodeExecutionId().GetExecutionId().GetProject(), request.GetEvent().GetParentNodeExecutionId().GetExecutionId().GetDomain(),
			request.GetEvent().GetParentNodeExecutionId().GetExecutionId().GetName(), request.GetEvent().GetParentNodeExecutionId().GetNodeId(),
			request.GetEvent().GetTaskId().GetProject(), request.GetEvent().GetTaskId().GetDomain(), request.GetEvent().GetTaskId().GetName(), request.GetEvent().GetTaskId().GetVersion(),
			strconv.FormatUint(uint64(request.GetEvent().GetRetryAttempt()), 10), InputsObjectSuffix)
		if err != nil {
			return err
		}
		logger.Debugf(ctx, "offloaded task execution inputs to [%s]", uri)
		taskExecutionModel.InputURI = uri.String()
	default:
		logger.Debugf(ctx, "request contained no input data")

	}
	return nil
}
