package impl

import (
	"context"
	"strconv"

	"github.com/golang/protobuf/proto"
	notificationInterfaces "github.com/lyft/flyteadmin/pkg/async/notifications/interfaces"

	"github.com/lyft/flytestdlib/storage"

	"github.com/lyft/flytestdlib/contextutils"

	"github.com/lyft/flyteadmin/pkg/manager/impl/shared"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/lyft/flytestdlib/logger"

	"github.com/lyft/flyteadmin/pkg/common"
	"github.com/lyft/flyteadmin/pkg/manager/impl/validation"

	"fmt"

	dataInterfaces "github.com/lyft/flyteadmin/pkg/data/interfaces"
	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/manager/impl/util"
	"github.com/lyft/flyteadmin/pkg/manager/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories"
	repoInterfaces "github.com/lyft/flyteadmin/pkg/repositories/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flyteadmin/pkg/repositories/transformers"
	runtimeInterfaces "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"google.golang.org/grpc/codes"
)

type nodeExecutionMetrics struct {
	Scope                      promutils.Scope
	ActiveNodeExecutions       prometheus.Gauge
	NodeExecutionsCreated      prometheus.Counter
	NodeExecutionsTerminated   prometheus.Counter
	NodeExecutionEventsCreated prometheus.Counter
	MissingWorkflowExecution   prometheus.Counter
	ClosureSizeBytes           prometheus.Summary
	NodeExecutionInputBytes    prometheus.Summary
	NodeExecutionOutputBytes   prometheus.Summary
	PublishEventError          prometheus.Counter
}

type NodeExecutionManager struct {
	db             repositories.RepositoryInterface
	config         runtimeInterfaces.Configuration
	storageClient  *storage.DataStore
	metrics        nodeExecutionMetrics
	urlData        dataInterfaces.RemoteURLInterface
	eventPublisher notificationInterfaces.Publisher
}

type updateNodeExecutionStatus int

const (
	updateSucceeded updateNodeExecutionStatus = iota
	updateFailed
	alreadyInTerminalStatus
)

var isParent = common.NewMapFilter(map[string]interface{}{
	shared.ParentTaskExecutionID: nil,
	shared.ParentID:              nil,
})

func getNodeExecutionContext(ctx context.Context, identifier *core.NodeExecutionIdentifier) context.Context {
	ctx = contextutils.WithProjectDomain(ctx, identifier.ExecutionId.Project, identifier.ExecutionId.Domain)
	ctx = contextutils.WithExecutionID(ctx, identifier.ExecutionId.Name)
	return contextutils.WithNodeID(ctx, identifier.NodeId)
}

func (m *NodeExecutionManager) createNodeExecutionWithEvent(
	ctx context.Context, request *admin.NodeExecutionEventRequest) error {

	var parentTaskExecutionID uint
	if request.Event.ParentTaskMetadata != nil {
		taskExecutionModel, err := util.GetTaskExecutionModel(ctx, m.db, request.Event.ParentTaskMetadata.Id)
		if err != nil {
			return err
		}
		parentTaskExecutionID = taskExecutionModel.ID
	}
	var parentID *uint
	if request.Event.ParentNodeMetadata != nil {
		parentNodeExecutionModel, err := util.GetNodeExecutionModel(ctx, m.db, &core.NodeExecutionIdentifier{
			ExecutionId: request.Event.Id.ExecutionId,
			NodeId:      request.Event.ParentNodeMetadata.NodeId,
		})
		if err != nil {
			logger.Errorf(ctx, "failed to fetch node execution for the parent node: %v %s with err",
				request.Event.Id.ExecutionId, request.Event.ParentNodeMetadata.NodeId, err)
			return err
		}
		parentID = &parentNodeExecutionModel.ID
	}
	nodeExecutionModel, err := transformers.CreateNodeExecutionModel(transformers.ToNodeExecutionModelInput{
		Request:               request,
		ParentTaskExecutionID: parentTaskExecutionID,
		ParentID:              parentID,
	})
	if err != nil {
		logger.Debugf(ctx, "failed to create node execution model for event request: %s with err: %v",
			request.RequestId, err)
		return err
	}
	nodeExecutionEventModel, err := transformers.CreateNodeExecutionEventModel(*request)
	if err != nil {
		logger.Debugf(ctx, "failed to transform node execution event request: %s into model with err: %v",
			request.RequestId, err)
		return err
	}

	if err := m.db.NodeExecutionRepo().Create(ctx, nodeExecutionEventModel, nodeExecutionModel); err != nil {
		logger.Debugf(ctx, "Failed to create node execution with id [%+v] and model [%+v] "+
			"and event [%+v] with err %v", request.Event.Id, nodeExecutionModel, nodeExecutionEventModel, err)
		return err
	}
	m.metrics.ClosureSizeBytes.Observe(float64(len(nodeExecutionModel.Closure)))
	return nil
}

func (m *NodeExecutionManager) updateNodeExecutionWithEvent(
	ctx context.Context, request *admin.NodeExecutionEventRequest, nodeExecutionModel *models.NodeExecution) (updateNodeExecutionStatus, error) {
	// If we have an existing execution, check if the phase change is valid
	nodeExecPhase := core.NodeExecution_Phase(core.NodeExecution_Phase_value[nodeExecutionModel.Phase])
	if nodeExecPhase == request.Event.Phase {
		logger.Debugf(ctx, "This phase was already recorded %v for %+v", nodeExecPhase.String(), request.Event.Id)
		return updateFailed, errors.NewFlyteAdminErrorf(codes.AlreadyExists,
			"This phase was already recorded %v for %+v", nodeExecPhase.String(), request.Event.Id)
	} else if common.IsNodeExecutionTerminal(nodeExecPhase) {
		// Cannot go from a terminal state to anything else
		logger.Warnf(ctx, "Invalid phase change from %v to %v for node execution %v",
			nodeExecPhase.String(), request.Event.Phase.String(), request.Event.Id)
		return alreadyInTerminalStatus, nil
	}

	// if this node execution kicked off a workflow, validate that the execution exists
	var childExecutionID *core.WorkflowExecutionIdentifier
	if request.Event.GetWorkflowNodeMetadata() != nil {
		childExecutionID = request.Event.GetWorkflowNodeMetadata().ExecutionId
		err := validation.ValidateWorkflowExecutionIdentifier(childExecutionID)
		if err != nil {
			logger.Errorf(ctx, "Invalid execution ID: %s with err: %v",
				childExecutionID, err)
		}
		_, err = util.GetExecutionModel(ctx, m.db, *childExecutionID)
		if err != nil {
			logger.Errorf(ctx, "The node execution launched an execution but it does not exist: %s with err: %v",
				childExecutionID, err)
			return updateFailed, err
		}
	}
	err := transformers.UpdateNodeExecutionModel(request, nodeExecutionModel, childExecutionID)
	if err != nil {
		logger.Debugf(ctx, "failed to update node execution model: %+v with err: %v", request.Event.Id, err)
		return updateFailed, err
	}

	nodeExecutionEventModel, err := transformers.CreateNodeExecutionEventModel(*request)
	if err != nil {
		logger.Debugf(ctx, "failed to create node execution event model for request: %s with err: %v",
			request.RequestId, err)
		return updateFailed, err
	}
	err = m.db.NodeExecutionRepo().Update(ctx, nodeExecutionEventModel, nodeExecutionModel)
	if err != nil {
		logger.Debugf(ctx, "Failed to update node execution with id [%+v] with err %v",
			request.Event.Id, err)
		return updateFailed, err
	}

	return updateSucceeded, nil
}

func (m *NodeExecutionManager) CreateNodeEvent(ctx context.Context, request admin.NodeExecutionEventRequest) (
	*admin.NodeExecutionEventResponse, error) {
	if err := validation.ValidateNodeExecutionIdentifier(request.Event.Id); err != nil {
		logger.Debugf(ctx, "CreateNodeEvent called with invalid identifier [%+v]: %v", request.Event.Id, err)
	}
	ctx = getNodeExecutionContext(ctx, request.Event.Id)
	executionID := request.Event.Id.ExecutionId
	logger.Debugf(ctx, "Received node execution event for Node Exec Id [%+v] transitioning to phase [%v], w/ Metadata [%v]",
		request.Event.Id, request.Event.Phase, request.Event.ParentTaskMetadata)

	_, err := util.GetExecutionModel(ctx, m.db, *executionID)
	if err != nil {
		m.metrics.MissingWorkflowExecution.Inc()
		logger.Debugf(ctx, "Failed to find existing execution with id [%+v] with err: %v", executionID, err)
		if ferr, ok := err.(errors.FlyteAdminError); ok {
			return nil, errors.NewFlyteAdminErrorf(ferr.Code(),
				"Failed to get existing execution id:[%+v] with err: %v", executionID, err)
		}
		return nil, fmt.Errorf("failed to get existing execution id: [%+v] with err: %v", executionID, err)
	}

	nodeExecutionModel, err := m.db.NodeExecutionRepo().Get(ctx, repoInterfaces.GetNodeExecutionInput{
		NodeExecutionIdentifier: *request.Event.Id,
	})
	if err != nil {
		if err.(errors.FlyteAdminError).Code() != codes.NotFound {
			logger.Debugf(ctx, "Failed to retrieve existing node execution with id [%+v] with err: %v",
				request.Event.Id, err)
			return nil, err
		}
		err = m.createNodeExecutionWithEvent(ctx, &request)
		if err != nil {
			return nil, err
		}
		m.metrics.NodeExecutionsCreated.Inc()
	} else {
		phase := core.NodeExecution_Phase(core.NodeExecution_Phase_value[nodeExecutionModel.Phase])
		updateStatus, err := m.updateNodeExecutionWithEvent(ctx, &request, &nodeExecutionModel)
		if err != nil {
			return nil, err
		}

		if updateStatus == alreadyInTerminalStatus {
			curPhase := request.Event.Phase.String()
			errorMsg := fmt.Sprintf("Invalid phase change from %s to %s for node execution %v", phase.String(), curPhase, nodeExecutionModel.ID)
			return nil, errors.NewAlreadyInTerminalStateError(ctx, errorMsg, curPhase)
		}
	}

	if request.Event.Phase == core.NodeExecution_RUNNING {
		m.metrics.ActiveNodeExecutions.Inc()
	} else if common.IsNodeExecutionTerminal(request.Event.Phase) {
		m.metrics.ActiveNodeExecutions.Dec()
		m.metrics.NodeExecutionsTerminated.Inc()
	}
	m.metrics.NodeExecutionEventsCreated.Inc()

	if err := m.eventPublisher.Publish(ctx, proto.MessageName(&request), &request); err != nil {
		m.metrics.PublishEventError.Inc()
		logger.Infof(ctx, "error publishing event [%+v] with err: [%v]", request.RequestId, err)
	}

	return &admin.NodeExecutionEventResponse{}, nil
}

func (m *NodeExecutionManager) GetNodeExecution(
	ctx context.Context, request admin.NodeExecutionGetRequest) (*admin.NodeExecution, error) {
	if err := validation.ValidateNodeExecutionIdentifier(request.Id); err != nil {
		logger.Debugf(ctx, "get node execution called with invalid identifier [%+v]: %v", request.Id, err)
	}
	ctx = getNodeExecutionContext(ctx, request.Id)
	nodeExecutionModel, err := util.GetNodeExecutionModel(ctx, m.db, request.Id)
	if err != nil {
		logger.Debugf(ctx, "Failed to get node execution with id [%+v] with err %v",
			request.Id, err)
		return nil, err
	}
	nodeExecution, err := transformers.FromNodeExecutionModel(*nodeExecutionModel)
	if err != nil {
		logger.Debugf(ctx, "failed to transform node execution model [%+v] to proto with err: %v", request.Id, err)
		return nil, err
	}
	return nodeExecution, nil
}

func (m *NodeExecutionManager) listNodeExecutions(
	ctx context.Context, identifierFilters []common.InlineFilter,
	requestFilters string, limit uint32, requestToken string, sortBy *admin.Sort, mapFilters []common.MapFilter) (
	*admin.NodeExecutionList, error) {

	filters, err := util.AddRequestFilters(requestFilters, common.NodeExecution, identifierFilters)
	if err != nil {
		return nil, err
	}
	var sortParameter common.SortParameter
	if sortBy != nil {
		sortParameter, err = common.NewSortParameter(*sortBy)
		if err != nil {
			return nil, err
		}
	}
	offset, err := validation.ValidateToken(requestToken)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"invalid pagination token %s for ListNodeExecutions", requestToken)
	}
	listInput := repoInterfaces.ListResourceInput{
		Limit:         int(limit),
		Offset:        offset,
		InlineFilters: filters,
		SortParameter: sortParameter,
	}

	listInput.MapFilters = mapFilters
	output, err := m.db.NodeExecutionRepo().List(ctx, listInput)
	if err != nil {
		logger.Debugf(ctx, "Failed to list node executions for request with err %v", err)
		return nil, err
	}

	var token string
	if len(output.NodeExecutions) == int(limit) {
		token = strconv.Itoa(offset + len(output.NodeExecutions))
	}
	nodeExecutionList, err := transformers.FromNodeExecutionModels(output.NodeExecutions)
	if err != nil {
		logger.Debugf(ctx, "failed to transform node execution models for request with err: %v", err)
		return nil, err
	}

	return &admin.NodeExecutionList{
		NodeExecutions: nodeExecutionList,
		Token:          token,
	}, nil
}

func (m *NodeExecutionManager) ListNodeExecutions(
	ctx context.Context, request admin.NodeExecutionListRequest) (*admin.NodeExecutionList, error) {
	// Check required fields
	if err := validation.ValidateNodeExecutionListRequest(request); err != nil {
		return nil, err
	}
	ctx = getExecutionContext(ctx, request.WorkflowExecutionId)

	identifierFilters, err := util.GetWorkflowExecutionIdentifierFilters(ctx, *request.WorkflowExecutionId)
	if err != nil {
		return nil, err
	}
	var mapFilters []common.MapFilter
	if request.UniqueParentId != "" {
		parentNodeExecution, err := util.GetNodeExecutionModel(ctx, m.db, &core.NodeExecutionIdentifier{
			ExecutionId: request.WorkflowExecutionId,
			NodeId:      request.UniqueParentId,
		})
		if err != nil {
			return nil, err
		}
		parentIDFilter, err := common.NewSingleValueFilter(
			common.NodeExecution, common.Equal, shared.ParentID, parentNodeExecution.ID)
		if err != nil {
			return nil, err
		}
		identifierFilters = append(identifierFilters, parentIDFilter)
	} else {
		mapFilters = []common.MapFilter{
			isParent,
		}
	}
	return m.listNodeExecutions(
		ctx, identifierFilters, request.Filters, request.Limit, request.Token, request.SortBy, mapFilters)
}

// Filters on node executions matching the execution parameters (execution project, domain, and name) as well as the
// parent task execution id corresponding to the task execution identified in the request params.
func (m *NodeExecutionManager) ListNodeExecutionsForTask(
	ctx context.Context, request admin.NodeExecutionForTaskListRequest) (*admin.NodeExecutionList, error) {
	// Check required fields
	if err := validation.ValidateNodeExecutionForTaskListRequest(request); err != nil {
		return nil, err
	}
	ctx = getTaskExecutionContext(ctx, request.TaskExecutionId)
	identifierFilters, err := util.GetWorkflowExecutionIdentifierFilters(
		ctx, *request.TaskExecutionId.NodeExecutionId.ExecutionId)
	if err != nil {
		return nil, err
	}
	parentTaskExecutionModel, err := util.GetTaskExecutionModel(ctx, m.db, request.TaskExecutionId)
	if err != nil {
		return nil, err
	}
	nodeIDFilter, err := common.NewSingleValueFilter(
		common.NodeExecution, common.Equal, shared.ParentTaskExecutionID, parentTaskExecutionModel.ID)
	if err != nil {
		return nil, err
	}
	identifierFilters = append(identifierFilters, nodeIDFilter)
	return m.listNodeExecutions(
		ctx, identifierFilters, request.Filters, request.Limit, request.Token, request.SortBy, nil)
}

func (m *NodeExecutionManager) GetNodeExecutionData(
	ctx context.Context, request admin.NodeExecutionGetDataRequest) (*admin.NodeExecutionGetDataResponse, error) {
	if err := validation.ValidateNodeExecutionIdentifier(request.Id); err != nil {
		logger.Debugf(ctx, "can't get node execution data with invalid identifier [%+v]: %v", request.Id, err)
	}
	ctx = getNodeExecutionContext(ctx, request.Id)
	nodeExecutionModel, err := util.GetNodeExecutionModel(ctx, m.db, request.Id)
	if err != nil {
		logger.Debugf(ctx, "Failed to get node execution with id [%+v] with err %v",
			request.Id, err)
		return nil, err
	}
	nodeExecution, err := transformers.FromNodeExecutionModel(*nodeExecutionModel)
	if err != nil {
		logger.Debugf(ctx, "failed to transform node execution model [%+v] when fetching data: %v", request.Id, err)
		return nil, err
	}
	signedInputsURLBlob, err := m.urlData.Get(ctx, nodeExecution.InputUri)
	if err != nil {
		return nil, err
	}
	signedOutputsURLBlob := admin.UrlBlob{}
	if nodeExecution.Closure.GetOutputUri() != "" {
		signedOutputsURLBlob, err = m.urlData.Get(ctx, nodeExecution.Closure.GetOutputUri())
		if err != nil {
			return nil, err
		}
	}
	response := &admin.NodeExecutionGetDataResponse{
		Inputs:  &signedInputsURLBlob,
		Outputs: &signedOutputsURLBlob,
	}
	maxDataSize := m.config.ApplicationConfiguration().GetRemoteDataConfig().MaxSizeInBytes
	if maxDataSize == 0 || signedInputsURLBlob.Bytes < maxDataSize {
		var fullInputs core.LiteralMap
		err := m.storageClient.ReadProtobuf(ctx, storage.DataReference(nodeExecution.InputUri), &fullInputs)
		if err != nil {
			logger.Warningf(ctx, "Failed to read inputs from URI [%s] with err: %v", nodeExecution.InputUri, err)
		}
		response.FullInputs = &fullInputs
	}
	if maxDataSize == 0 || (signedOutputsURLBlob.Bytes < maxDataSize && len(nodeExecution.Closure.GetOutputUri()) > 0) {
		var fullOutputs core.LiteralMap
		err := m.storageClient.ReadProtobuf(ctx, storage.DataReference(nodeExecution.Closure.GetOutputUri()), &fullOutputs)
		if err != nil {
			logger.Warningf(ctx, "Failed to read outputs from URI [%s] with err: %v",
				nodeExecution.Closure.GetOutputUri(), err)
		}
		response.FullOutputs = &fullOutputs
	}

	m.metrics.NodeExecutionInputBytes.Observe(float64(response.Inputs.Bytes))
	m.metrics.NodeExecutionOutputBytes.Observe(float64(response.Outputs.Bytes))

	return response, nil
}

func NewNodeExecutionManager(db repositories.RepositoryInterface, config runtimeInterfaces.Configuration, storageClient *storage.DataStore, scope promutils.Scope, urlData dataInterfaces.RemoteURLInterface, eventPublisher notificationInterfaces.Publisher) interfaces.NodeExecutionInterface {
	metrics := nodeExecutionMetrics{
		Scope: scope,
		ActiveNodeExecutions: scope.MustNewGauge("active_node_executions",
			"overall count of active node executions"),
		NodeExecutionsCreated: scope.MustNewCounter("node_executions_created",
			"overall count of node executions created"),
		NodeExecutionsTerminated: scope.MustNewCounter("node_executions_terminated",
			"overall count of terminated node executions"),
		NodeExecutionEventsCreated: scope.MustNewCounter("node_execution_events_created",
			"overall count of successfully completed NodeExecutionEventRequest"),
		MissingWorkflowExecution: scope.MustNewCounter("missing_workflow_execution",
			"overall count of node execution events received that are missing a parent workflow execution"),
		ClosureSizeBytes: scope.MustNewSummary("closure_size_bytes",
			"size in bytes of serialized node execution closure"),
		NodeExecutionInputBytes: scope.MustNewSummary("input_size_bytes",
			"size in bytes of serialized node execution inputs"),
		NodeExecutionOutputBytes: scope.MustNewSummary("output_size_bytes",
			"size in bytes of serialized node execution outputs"),
		PublishEventError: scope.MustNewCounter("publish_event_error",
			"overall count of publish event errors when invoking publish()"),
	}
	return &NodeExecutionManager{
		db:             db,
		config:         config,
		storageClient:  storageClient,
		metrics:        metrics,
		urlData:        urlData,
		eventPublisher: eventPublisher,
	}
}
