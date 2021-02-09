package impl

import (
	"context"
	"fmt"
	"strconv"

	"github.com/golang/protobuf/proto"
	notificationInterfaces "github.com/lyft/flyteadmin/pkg/async/notifications/interfaces"

	"github.com/lyft/flytestdlib/storage"

	"github.com/lyft/flytestdlib/contextutils"

	"github.com/lyft/flytestdlib/promutils"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/lyft/flyteadmin/pkg/common"
	dataInterfaces "github.com/lyft/flyteadmin/pkg/data/interfaces"
	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/manager/impl/util"
	"github.com/lyft/flyteadmin/pkg/manager/impl/validation"
	"github.com/lyft/flyteadmin/pkg/manager/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories"
	repoInterfaces "github.com/lyft/flyteadmin/pkg/repositories/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flyteadmin/pkg/repositories/transformers"
	runtimeInterfaces "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytestdlib/logger"
	"google.golang.org/grpc/codes"
)

type taskExecutionMetrics struct {
	Scope                      promutils.Scope
	ActiveTaskExecutions       prometheus.Gauge
	TaskExecutionsCreated      prometheus.Counter
	TaskExecutionsTerminated   prometheus.Counter
	TaskExecutionEventsCreated prometheus.Counter
	MissingTaskExecution       prometheus.Counter
	MissingTaskDefinition      prometheus.Counter
	ClosureSizeBytes           prometheus.Summary
	TaskExecutionInputBytes    prometheus.Summary
	TaskExecutionOutputBytes   prometheus.Summary
	PublishEventError          prometheus.Counter
}

type TaskExecutionManager struct {
	db                 repositories.RepositoryInterface
	config             runtimeInterfaces.Configuration
	storageClient      *storage.DataStore
	metrics            taskExecutionMetrics
	urlData            dataInterfaces.RemoteURLInterface
	notificationClient notificationInterfaces.Publisher
}

func getTaskExecutionContext(ctx context.Context, identifier *core.TaskExecutionIdentifier) context.Context {
	ctx = getNodeExecutionContext(ctx, identifier.NodeExecutionId)
	return contextutils.WithTaskID(ctx, fmt.Sprintf("%s-%v", identifier.TaskId.Name, identifier.RetryAttempt))
}

func (m *TaskExecutionManager) createTaskExecution(
	ctx context.Context, nodeExecutionModel *models.NodeExecution, request *admin.TaskExecutionEventRequest) (
	models.TaskExecution, error) {

	taskExecutionModel, err := transformers.CreateTaskExecutionModel(
		transformers.CreateTaskExecutionModelInput{
			Request: request,
		})
	if err != nil {
		logger.Debugf(ctx, "failed to transform task execution %+v into database model: %v", request.Event.TaskId, err)
		return models.TaskExecution{}, err
	}
	if err := m.db.TaskExecutionRepo().Create(ctx, *taskExecutionModel); err != nil {
		logger.Debugf(ctx, "Failed to create task execution with task id [%+v] and node execution model [%+v] with err %v",
			request.Event.TaskId, nodeExecutionModel, err)
		return models.TaskExecution{}, err
	}

	m.metrics.TaskExecutionsCreated.Inc()
	m.metrics.ClosureSizeBytes.Observe(float64(len(nodeExecutionModel.Closure)))
	logger.Debugf(ctx, "created task execution: %+v", request.Event.TaskId)
	return *taskExecutionModel, nil
}

func (m *TaskExecutionManager) updateTaskExecutionModelState(
	ctx context.Context, request *admin.TaskExecutionEventRequest, existingTaskExecution *models.TaskExecution) (
	models.TaskExecution, error) {

	err := transformers.UpdateTaskExecutionModel(request, existingTaskExecution)
	if err != nil {
		logger.Debugf(ctx, "failed to update task execution model [%+v] with err: %v", request.Event.TaskId, err)
		return models.TaskExecution{}, err
	}

	err = m.db.TaskExecutionRepo().Update(ctx, *existingTaskExecution)
	if err != nil {
		logger.Debugf(ctx, "Failed to update task execution with task id [%+v] and task execution model [%+v] with err %v",
			request.Event.TaskId, existingTaskExecution, err)
		return models.TaskExecution{}, err
	}

	return *existingTaskExecution, nil
}

func (m *TaskExecutionManager) CreateTaskExecutionEvent(ctx context.Context, request admin.TaskExecutionEventRequest) (
	*admin.TaskExecutionEventResponse, error) {
	// Get the parent node execution, if none found a MissingEntityError will be returned
	nodeExecutionID := request.Event.ParentNodeExecutionId
	taskExecutionID := core.TaskExecutionIdentifier{
		TaskId:          request.Event.TaskId,
		NodeExecutionId: nodeExecutionID,
		RetryAttempt:    request.Event.RetryAttempt,
	}
	ctx = getTaskExecutionContext(ctx, &taskExecutionID)
	logger.Debugf(ctx, "Received task execution event for [%+v] transitioning to phase [%v]",
		taskExecutionID, request.Event.Phase)
	nodeExecutionModel, err := util.GetNodeExecutionModel(ctx, m.db, nodeExecutionID)
	if err != nil {
		m.metrics.MissingTaskExecution.Inc()
		logger.Debugf(ctx, "Failed to get existing node execution [%+v] with err %v", nodeExecutionID, err)
		if ferr, ok := err.(errors.FlyteAdminError); ok {
			return nil, errors.NewFlyteAdminErrorf(ferr.Code(),
				"Failed to get existing execution node id:[%+v] with err: %v", nodeExecutionID, err)
		}
		return nil, fmt.Errorf("failed to get existing node execution id: [%+v] with err: %v", nodeExecutionID, err)
	}

	// See if the task execution exists
	// - if it does check if the new phase is applicable and then update
	// - if it doesn't, create a task execution
	taskExecutionModel, err := m.db.TaskExecutionRepo().Get(ctx, repoInterfaces.GetTaskExecutionInput{
		TaskExecutionID: taskExecutionID,
	})

	if err != nil {
		if err.(errors.FlyteAdminError).Code() != codes.NotFound {
			logger.Debugf(ctx, "Failed to find existing task execution [%+v] with err %v", taskExecutionID, err)
			return nil, err
		}
		_, err := m.createTaskExecution(ctx, nodeExecutionModel, &request)
		if err != nil {
			return nil, err
		}

		return &admin.TaskExecutionEventResponse{}, nil
	}
	if taskExecutionModel.Phase == request.Event.Phase.String() &&
		taskExecutionModel.PhaseVersion >= request.Event.PhaseVersion {
		logger.Debugf(ctx, "have already recorded task execution phase %s (version: %d) for %v",
			request.Event.Phase.String(), request.Event.PhaseVersion, taskExecutionID)
		return nil, errors.NewFlyteAdminErrorf(codes.AlreadyExists,
			"have already recorded task execution phase %s (version: %d) for %v",
			request.Event.Phase.String(), request.Event.PhaseVersion, taskExecutionID)
	}

	currentPhase := core.TaskExecution_Phase(core.TaskExecution_Phase_value[taskExecutionModel.Phase])
	if common.IsTaskExecutionTerminal(currentPhase) {
		// Cannot update a terminal execution.
		curPhase := request.Event.Phase.String()
		errorMsg := fmt.Sprintf("invalid phase change from %v to %v for task execution %v", taskExecutionModel.Phase, request.Event.Phase, taskExecutionID)
		logger.Warnf(ctx, errorMsg)
		return nil, errors.NewAlreadyInTerminalStateError(ctx, errorMsg, curPhase)
	}

	taskExecutionModel, err = m.updateTaskExecutionModelState(ctx, &request, &taskExecutionModel)
	if err != nil {
		logger.Debugf(ctx, "Failed to update task execution with id [%+v] with err %v",
			taskExecutionID, err)
		return nil, err
	}

	if request.Event.Phase == core.TaskExecution_RUNNING && request.Event.PhaseVersion == 0 {
		m.metrics.ActiveTaskExecutions.Inc()
	} else if common.IsTaskExecutionTerminal(request.Event.Phase) && request.Event.PhaseVersion == 0 {
		m.metrics.ActiveTaskExecutions.Dec()
		m.metrics.TaskExecutionsTerminated.Inc()
	}

	if err = m.notificationClient.Publish(ctx, proto.MessageName(&request), &request); err != nil {
		m.metrics.PublishEventError.Inc()
		logger.Infof(ctx, "error publishing event [%+v] with err: [%v]", request.RequestId, err)
	}

	m.metrics.TaskExecutionEventsCreated.Inc()
	logger.Debugf(ctx, "Successfully recorded task execution event [%v]", request.Event)
	// TODO: we will want to return some scope information here soon!
	return &admin.TaskExecutionEventResponse{}, nil
}

func (m *TaskExecutionManager) GetTaskExecution(
	ctx context.Context, request admin.TaskExecutionGetRequest) (*admin.TaskExecution, error) {
	err := validation.ValidateTaskExecutionIdentifier(request.Id)
	if err != nil {
		logger.Debugf(ctx, "Failed to validate GetTaskExecution [%+v] with err: %v", request.Id, err)
		return nil, err
	}
	ctx = getTaskExecutionContext(ctx, request.Id)
	taskExecutionModel, err := util.GetTaskExecutionModel(ctx, m.db, request.Id)
	if err != nil {
		return nil, err
	}
	taskExecution, err := transformers.FromTaskExecutionModel(*taskExecutionModel)
	if err != nil {
		logger.Debugf(ctx, "Failed to transform task execution model [%+v] to proto: %v", request.Id, err)
		return nil, err
	}
	return taskExecution, nil
}

func (m *TaskExecutionManager) ListTaskExecutions(
	ctx context.Context, request admin.TaskExecutionListRequest) (*admin.TaskExecutionList, error) {
	if err := validation.ValidateTaskExecutionListRequest(request); err != nil {
		logger.Debugf(ctx, "ListTaskExecutions request [%+v] is invalid: %v", request, err)
		return nil, err
	}
	ctx = getNodeExecutionContext(ctx, request.NodeExecutionId)

	identifierFilters, err := util.GetNodeExecutionIdentifierFilters(ctx, *request.NodeExecutionId)
	if err != nil {
		return nil, err
	}

	filters, err := util.AddRequestFilters(request.Filters, common.TaskExecution, identifierFilters)
	if err != nil {
		return nil, err
	}
	var sortParameter common.SortParameter
	if request.SortBy != nil {
		sortParameter, err = common.NewSortParameter(*request.SortBy)
		if err != nil {
			return nil, err
		}
	}

	offset, err := validation.ValidateToken(request.Token)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"invalid pagination token %s for ListTaskExecutions", request.Token)
	}

	output, err := m.db.TaskExecutionRepo().List(ctx, repoInterfaces.ListResourceInput{
		InlineFilters: filters,
		Offset:        offset,
		Limit:         int(request.Limit),
		SortParameter: sortParameter,
	})
	if err != nil {
		logger.Debugf(ctx, "Failed to list task executions with request [%+v] with err %v",
			request, err)
		return nil, err
	}

	taskExecutionList, err := transformers.FromTaskExecutionModels(output.TaskExecutions)
	if err != nil {
		logger.Debugf(ctx, "failed to transform task execution models for request [%+v] with err: %v", request, err)
		return nil, err
	}
	var token string
	if len(taskExecutionList) == int(request.Limit) {
		token = strconv.Itoa(offset + len(taskExecutionList))
	}
	return &admin.TaskExecutionList{
		TaskExecutions: taskExecutionList,
		Token:          token,
	}, nil
}

func (m *TaskExecutionManager) GetTaskExecutionData(
	ctx context.Context, request admin.TaskExecutionGetDataRequest) (*admin.TaskExecutionGetDataResponse, error) {
	if err := validation.ValidateTaskExecutionIdentifier(request.Id); err != nil {
		logger.Debugf(ctx, "Invalid identifier [%+v]: %v", request.Id, err)
	}
	ctx = getTaskExecutionContext(ctx, request.Id)
	taskExecution, err := m.GetTaskExecution(ctx, admin.TaskExecutionGetRequest{
		Id: request.Id,
	})
	if err != nil {
		logger.Debugf(ctx, "Failed to get task execution with id [%+v] with err %v",
			request.Id, err)
		return nil, err
	}
	signedInputsURLBlob, err := m.urlData.Get(ctx, taskExecution.InputUri)
	if err != nil {
		return nil, err
	}
	signedOutputsURLBlob := admin.UrlBlob{}
	if taskExecution.Closure.GetOutputUri() != "" {
		signedOutputsURLBlob, err = m.urlData.Get(ctx, taskExecution.Closure.GetOutputUri())
		if err != nil {
			return nil, err
		}
	}
	response := &admin.TaskExecutionGetDataResponse{
		Inputs:  &signedInputsURLBlob,
		Outputs: &signedOutputsURLBlob,
	}
	maxDataSize := m.config.ApplicationConfiguration().GetRemoteDataConfig().MaxSizeInBytes
	if maxDataSize == 0 || signedInputsURLBlob.Bytes < maxDataSize {
		var fullInputs core.LiteralMap
		err := m.storageClient.ReadProtobuf(ctx, storage.DataReference(taskExecution.InputUri), &fullInputs)
		if err != nil {
			logger.Warningf(ctx, "Failed to read inputs from URI [%s] with err: %v", taskExecution.InputUri, err)
		}
		response.FullInputs = &fullInputs
	}
	if maxDataSize == 0 || (signedOutputsURLBlob.Bytes < maxDataSize && len(taskExecution.Closure.GetOutputUri()) > 0) {
		var fullOutputs core.LiteralMap
		err := m.storageClient.ReadProtobuf(ctx, storage.DataReference(taskExecution.Closure.GetOutputUri()), &fullOutputs)
		if err != nil {
			logger.Warningf(ctx, "Failed to read outputs from URI [%s] with err: %v",
				taskExecution.Closure.GetOutputUri(), err)
		}
		response.FullOutputs = &fullOutputs
	}

	m.metrics.TaskExecutionInputBytes.Observe(float64(response.Inputs.Bytes))
	m.metrics.TaskExecutionOutputBytes.Observe(float64(response.Outputs.Bytes))
	return response, nil
}

func NewTaskExecutionManager(db repositories.RepositoryInterface, config runtimeInterfaces.Configuration, storageClient *storage.DataStore, scope promutils.Scope, urlData dataInterfaces.RemoteURLInterface, publisher notificationInterfaces.Publisher) interfaces.TaskExecutionInterface {
	metrics := taskExecutionMetrics{
		Scope: scope,
		ActiveTaskExecutions: scope.MustNewGauge("active_executions",
			"overall count of active task executions"),
		MissingTaskExecution: scope.MustNewCounter("missing_node_execution",
			"overall count of task execution events received that are missing a parent node execution"),
		TaskExecutionsCreated: scope.MustNewCounter("task_executions_created",
			"overall count of successfully completed CreateExecutionRequests"),
		TaskExecutionsTerminated: scope.MustNewCounter("task_executions_terminated",
			"overall count of terminated workflow executions"),
		TaskExecutionEventsCreated: scope.MustNewCounter("task_execution_events_created",
			"overall count of successfully completed WorkflowExecutionEventRequest"),
		MissingTaskDefinition: scope.MustNewCounter("missing_task_definition",
			"overall count of task execution events received that are missing a task definition"),
		ClosureSizeBytes: scope.MustNewSummary("closure_size_bytes",
			"size in bytes of serialized task execution closure"),
		TaskExecutionInputBytes: scope.MustNewSummary("input_size_bytes",
			"size in bytes of serialized node execution inputs"),
		TaskExecutionOutputBytes: scope.MustNewSummary("output_size_bytes",
			"size in bytes of serialized node execution outputs"),
		PublishEventError: scope.MustNewCounter("publish_event_error",
			"overall count of publish event errors when invoking publish()"),
	}
	return &TaskExecutionManager{
		db:                 db,
		config:             config,
		storageClient:      storageClient,
		metrics:            metrics,
		urlData:            urlData,
		notificationClient: publisher,
	}
}
