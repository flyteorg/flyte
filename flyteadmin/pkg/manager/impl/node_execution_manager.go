package impl

import (
	"context"
	"fmt"
	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"

	cloudeventInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/async/cloudevent/interfaces"
	eventWriter "github.com/flyteorg/flyte/flyteadmin/pkg/async/events/interfaces"
	notificationInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/async/notifications/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	dataInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/data/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/util"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/validation"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/interfaces"
	repoInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/transformers"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

type nodeExecutionMetrics struct {
	Scope                      promutils.Scope
	ActiveNodeExecutions       prometheus.Gauge
	NodeExecutionsCreated      prometheus.Counter
	NodeExecutionsTerminated   labeled.Counter
	NodeExecutionEventsCreated prometheus.Counter
	MissingWorkflowExecution   prometheus.Counter
	ClosureSizeBytes           prometheus.Summary
	NodeExecutionInputBytes    prometheus.Summary
	NodeExecutionOutputBytes   prometheus.Summary
	PublishEventError          prometheus.Counter
}

type NodeExecutionManager struct {
	db                  repoInterfaces.Repository
	config              runtimeInterfaces.Configuration
	storagePrefix       []string
	storageClient       *storage.DataStore
	metrics             nodeExecutionMetrics
	urlData             dataInterfaces.RemoteURLInterface
	eventPublisher      notificationInterfaces.Publisher
	cloudEventPublisher cloudeventInterfaces.Publisher
	dbEventWriter       eventWriter.NodeExecutionEventWriter
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
	ctx = contextutils.WithProjectDomain(ctx, identifier.GetExecutionId().GetProject(), identifier.GetExecutionId().GetDomain())
	ctx = contextutils.WithExecutionID(ctx, identifier.GetExecutionId().GetName())
	return contextutils.WithNodeID(ctx, identifier.GetNodeId())
}

func (m *NodeExecutionManager) createNodeExecutionWithEvent(
	ctx context.Context, request *admin.NodeExecutionEventRequest, dynamicWorkflowRemoteClosureReference string) error {
	var parentTaskExecutionID *uint
	if request.GetEvent().GetParentTaskMetadata() != nil {
		taskExecutionModel, err := util.GetTaskExecutionModel(ctx, m.db, request.GetEvent().GetParentTaskMetadata().GetId())
		if err != nil {
			return err
		}
		parentTaskExecutionID = &taskExecutionModel.ID
	}
	var parentID *uint
	if request.GetEvent().GetParentNodeMetadata() != nil {
		parentNodeExecutionModel, err := util.GetNodeExecutionModel(ctx, m.db, &core.NodeExecutionIdentifier{
			ExecutionId: request.GetEvent().GetId().GetExecutionId(),
			NodeId:      request.GetEvent().GetParentNodeMetadata().GetNodeId(),
		})
		if err != nil {
			logger.Errorf(ctx, "failed to fetch node execution for the parent node: %v %s with err",
				request.GetEvent().GetId().GetExecutionId(), request.GetEvent().GetParentNodeMetadata().GetNodeId(), err)
			return err
		}
		parentID = &parentNodeExecutionModel.ID
	}
	nodeExecutionModel, err := transformers.CreateNodeExecutionModel(ctx, transformers.ToNodeExecutionModelInput{
		Request:                      request,
		ParentTaskExecutionID:        parentTaskExecutionID,
		ParentID:                     parentID,
		DynamicWorkflowRemoteClosure: dynamicWorkflowRemoteClosureReference,
		InlineEventDataPolicy:        m.config.ApplicationConfiguration().GetRemoteDataConfig().InlineEventDataPolicy,
		StorageClient:                m.storageClient,
	})
	if err != nil {
		logger.Debugf(ctx, "failed to create node execution model for event request: %s with err: %v",
			request.GetRequestId(), err)
		return err
	}
	if err := m.db.NodeExecutionRepo().Create(ctx, nodeExecutionModel); err != nil {
		logger.Debugf(ctx, "Failed to create node execution with id [%+v] and model [%+v] "+
			"with err %v", request.GetEvent().GetId(), nodeExecutionModel, err)
		return err
	}
	m.metrics.ClosureSizeBytes.Observe(float64(len(nodeExecutionModel.Closure)))
	return nil
}

func (m *NodeExecutionManager) updateNodeExecutionWithEvent(
	ctx context.Context, request *admin.NodeExecutionEventRequest, nodeExecutionModel *models.NodeExecution,
	dynamicWorkflowRemoteClosureReference string) (updateNodeExecutionStatus, error) {
	// If we have an existing execution, check if the phase change is valid
	nodeExecPhase := core.NodeExecution_Phase(core.NodeExecution_Phase_value[nodeExecutionModel.Phase])
	if nodeExecPhase == request.GetEvent().GetPhase() {
		logger.Debugf(ctx, "This phase was already recorded %v for %+v", nodeExecPhase.String(), request.GetEvent().GetId())
		return updateFailed, errors.NewFlyteAdminErrorf(codes.AlreadyExists,
			"This phase was already recorded %v for %+v", nodeExecPhase.String(), request.GetEvent().GetId())
	} else if common.IsNodeExecutionTerminal(nodeExecPhase) {
		// Cannot go from a terminal state to anything else
		logger.Warnf(ctx, "Invalid phase change from %v to %v for node execution %v",
			nodeExecPhase.String(), request.GetEvent().GetPhase().String(), request.GetEvent().GetId())
		return alreadyInTerminalStatus, nil
	}

	// if this node execution kicked off a workflow, validate that the execution exists
	var childExecutionID *core.WorkflowExecutionIdentifier
	if request.GetEvent().GetWorkflowNodeMetadata() != nil {
		childExecutionID = request.GetEvent().GetWorkflowNodeMetadata().GetExecutionId()
		err := validation.ValidateWorkflowExecutionIdentifier(childExecutionID)
		if err != nil {
			logger.Errorf(ctx, "Invalid execution ID: %s with err: %v",
				childExecutionID, err)
		}
		_, err = util.GetExecutionModel(ctx, m.db, childExecutionID)
		if err != nil {
			logger.Errorf(ctx, "The node execution launched an execution but it does not exist: %s with err: %v",
				childExecutionID, err)
			return updateFailed, err
		}
	}
	err := transformers.UpdateNodeExecutionModel(ctx, request, nodeExecutionModel, childExecutionID,
		dynamicWorkflowRemoteClosureReference, m.config.ApplicationConfiguration().GetRemoteDataConfig().InlineEventDataPolicy,
		m.storageClient)
	if err != nil {
		logger.Debugf(ctx, "failed to update node execution model: %+v with err: %v", request.GetEvent().GetId(), err)
		return updateFailed, err
	}
	err = m.db.NodeExecutionRepo().Update(ctx, nodeExecutionModel)
	if err != nil {
		logger.Debugf(ctx, "Failed to update node execution with id [%+v] with err %v",
			request.GetEvent().GetId(), err)
		return updateFailed, err
	}

	return updateSucceeded, nil
}

func formatDynamicWorkflowID(identifier *core.Identifier) string {
	return fmt.Sprintf("%s_%s_%s_%s", identifier.GetProject(), identifier.GetDomain(), identifier.GetName(), identifier.GetVersion())
}

func (m *NodeExecutionManager) uploadDynamicWorkflowClosure(
	ctx context.Context, nodeID *core.NodeExecutionIdentifier, workflowID *core.Identifier,
	compiledWorkflowClosure *core.CompiledWorkflowClosure) (storage.DataReference, error) {
	nestedSubKeys := []string{
		nodeID.GetExecutionId().GetProject(),
		nodeID.GetExecutionId().GetDomain(),
		nodeID.GetExecutionId().GetName(),
		nodeID.GetNodeId(),
		formatDynamicWorkflowID(workflowID),
	}
	nestedKeys := append(m.storagePrefix, nestedSubKeys...)
	remoteClosureDataRef, err := m.storageClient.ConstructReference(ctx, m.storageClient.GetBaseContainerFQN(ctx), nestedKeys...)

	if err != nil {
		return "", errors.NewFlyteAdminErrorf(codes.Internal,
			"Failed to produce remote closure data reference for dynamic workflow yielded by node id [%+v] with workflow id [%+v]; err: %v", nodeID, workflowID, err)
	}

	err = m.storageClient.WriteProtobuf(ctx, remoteClosureDataRef, defaultStorageOptions, compiledWorkflowClosure)
	if err != nil {
		return "", errors.NewFlyteAdminErrorf(codes.Internal,
			"Failed to upload dynamic workflow closure for node id [%+v] and workflow id [%+v] with err: %v", nodeID, workflowID, err)
	}
	return remoteClosureDataRef, nil
}

func (m *NodeExecutionManager) CreateNodeEvent(ctx context.Context, request *admin.NodeExecutionEventRequest) (
	*admin.NodeExecutionEventResponse, error) {
	if err := validation.ValidateNodeExecutionEventRequest(request, m.config.ApplicationConfiguration().GetRemoteDataConfig().MaxSizeInBytes); err != nil {
		logger.Debugf(ctx, "CreateNodeEvent called with invalid identifier [%+v]: %v", request.GetEvent().GetId(), err)
	}
	ctx = getNodeExecutionContext(ctx, request.GetEvent().GetId())
	logger.Debugf(ctx, "Received node execution event for Node Exec Id [%+v] transitioning to phase [%v], w/ Metadata [%v]",
		request.GetEvent().GetId(), request.GetEvent().GetPhase(), request.GetEvent().GetParentTaskMetadata())

	executionID := request.GetEvent().GetId().GetExecutionId()
	workflowExecution, err := m.db.ExecutionRepo().Get(ctx, repoInterfaces.Identifier{
		Project: executionID.GetProject(),
		Domain:  executionID.GetDomain(),
		Name:    executionID.GetName(),
	})
	if err != nil {
		m.metrics.MissingWorkflowExecution.Inc()
		logger.Debugf(ctx, "Failed to find existing execution with id [%+v] with err: %v", executionID, err)
		if err != nil {
			if ferr, ok := err.(errors.FlyteAdminError); ok {
				return nil, errors.NewFlyteAdminErrorf(ferr.Code(),
					"Failed to get existing execution id: [%+v] with err: %v", executionID, err)
			}
		}
		return nil, fmt.Errorf("failed to get existing execution id: [%+v]", executionID)
	}

	if err := validation.ValidateCluster(ctx, workflowExecution.Cluster, request.GetEvent().GetProducerId()); err != nil {
		return nil, err
	}

	var dynamicWorkflowRemoteClosureReference string
	if request.GetEvent().GetTaskNodeMetadata() != nil && request.GetEvent().GetTaskNodeMetadata().GetDynamicWorkflow() != nil {
		dynamicWorkflowRemoteClosureDataReference, err := m.uploadDynamicWorkflowClosure(
			ctx, request.GetEvent().GetId(), request.GetEvent().GetTaskNodeMetadata().GetDynamicWorkflow().GetId(),
			request.GetEvent().GetTaskNodeMetadata().GetDynamicWorkflow().GetCompiledWorkflow())
		if err != nil {
			return nil, err
		}
		dynamicWorkflowRemoteClosureReference = dynamicWorkflowRemoteClosureDataReference.String()
	}

	nodeExecutionModel, err := m.db.NodeExecutionRepo().Get(ctx, repoInterfaces.NodeExecutionResource{
		NodeExecutionIdentifier: request.GetEvent().GetId(),
	})
	if err != nil {
		if err.(errors.FlyteAdminError).Code() != codes.NotFound {
			logger.Debugf(ctx, "Failed to retrieve existing node execution with id [%+v] with err: %v",
				request.GetEvent().GetId(), err)
			return nil, err
		}
		err = m.createNodeExecutionWithEvent(ctx, request, dynamicWorkflowRemoteClosureReference)
		if err != nil {
			return nil, err
		}
		m.metrics.NodeExecutionsCreated.Inc()
	} else {
		phase := core.NodeExecution_Phase(core.NodeExecution_Phase_value[nodeExecutionModel.Phase])
		updateStatus, err := m.updateNodeExecutionWithEvent(ctx, request, &nodeExecutionModel, dynamicWorkflowRemoteClosureReference)
		if err != nil {
			return nil, err
		}

		if updateStatus == alreadyInTerminalStatus {
			curPhase := request.GetEvent().GetPhase().String()
			errorMsg := fmt.Sprintf("Invalid phase change from %s to %s for node execution %v", phase.String(), curPhase, nodeExecutionModel.ID)
			return nil, errors.NewAlreadyInTerminalStateError(ctx, errorMsg, curPhase)
		}
	}
	m.dbEventWriter.Write(request)

	if request.GetEvent().GetPhase() == core.NodeExecution_RUNNING {
		m.metrics.ActiveNodeExecutions.Inc()
	} else if common.IsNodeExecutionTerminal(request.GetEvent().GetPhase()) {
		m.metrics.ActiveNodeExecutions.Dec()
		m.metrics.NodeExecutionsTerminated.Inc(contextutils.WithPhase(ctx, request.GetEvent().GetPhase().String()))
		if request.GetEvent().GetOutputData() != nil {
			m.metrics.NodeExecutionOutputBytes.Observe(float64(proto.Size(request.GetEvent().GetOutputData())))
		}
	}
	m.metrics.NodeExecutionEventsCreated.Inc()

	if err := m.eventPublisher.Publish(ctx, proto.MessageName(request), request); err != nil {
		m.metrics.PublishEventError.Inc()
		logger.Infof(ctx, "error publishing event [%+v] with err: [%v]", request.GetRequestId(), err)
	}

	go func() {
		ceCtx := context.TODO()
		if err := m.cloudEventPublisher.Publish(ceCtx, proto.MessageName(request), request); err != nil {
			logger.Infof(ctx, "error publishing cloud event [%+v] with err: [%v]", request.GetRequestId(), err)
		}
	}()

	return &admin.NodeExecutionEventResponse{}, nil
}

func (m *NodeExecutionManager) GetDynamicNodeWorkflow(ctx context.Context, request *admin.GetDynamicNodeWorkflowRequest) (*admin.DynamicNodeWorkflowResponse, error) {
	if err := validation.ValidateNodeExecutionIdentifier(request.GetId()); err != nil {
		logger.Debugf(ctx, "can't get node execution data with invalid identifier [%+v]: %v", request.GetId(), err)
	}

	ctx = getNodeExecutionContext(ctx, request.GetId())
	nodeExecutionModel, err := util.GetNodeExecutionModel(ctx, m.db, request.GetId())
	if err != nil {
		logger.Errorf(ctx, "failed to get node execution with id [%+v] with err %v",
			request.GetId(), err)
		return nil, err
	}

	if nodeExecutionModel.DynamicWorkflowRemoteClosureReference == "" {
		return &admin.DynamicNodeWorkflowResponse{}, errors.NewFlyteAdminErrorf(codes.NotFound, "node does not contain dynamic workflow")
	}

	closure, err := m.fetchDynamicWorkflowClosure(ctx, nodeExecutionModel.DynamicWorkflowRemoteClosureReference)
	if err != nil {
		return nil, err
	}

	return &admin.DynamicNodeWorkflowResponse{CompiledWorkflow: closure}, nil
}

// Handles making additional database calls, if necessary, to populate IsParent & IsDynamic data using the historical pattern of
// preloading child node executions. Otherwise, simply calls transform on the input model.
func (m *NodeExecutionManager) transformNodeExecutionModel(ctx context.Context, nodeExecutionModel models.NodeExecution,
	nodeExecutionID *core.NodeExecutionIdentifier, opts *transformers.ExecutionTransformerOptions) (*admin.NodeExecution, error) {
	internalData, err := transformers.GetNodeExecutionInternalData(nodeExecutionModel.InternalData)
	if err != nil {
		return nil, err
	}
	if internalData.GetEventVersion() == 0 {
		// Issue more expensive query to determine whether this node is a parent and/or dynamic node.
		nodeExecutionModel, err = m.db.NodeExecutionRepo().GetWithChildren(ctx, repoInterfaces.NodeExecutionResource{
			NodeExecutionIdentifier: nodeExecutionID,
		})
		if err != nil {
			return nil, err
		}
	}

	nodeExecution, err := transformers.FromNodeExecutionModel(nodeExecutionModel, opts)
	if err != nil {
		logger.Debugf(ctx, "failed to transform node execution model [%+v] to proto with err: %v", nodeExecutionID, err)
		return nil, err
	}
	return nodeExecution, nil
}

func (m *NodeExecutionManager) transformNodeExecutionModelList(ctx context.Context, nodeExecutionModels []models.NodeExecution) ([]*admin.NodeExecution, error) {
	nodeExecutions := make([]*admin.NodeExecution, len(nodeExecutionModels))
	for idx, nodeExecutionModel := range nodeExecutionModels {
		nodeExecution, err := m.transformNodeExecutionModel(ctx, nodeExecutionModel, &core.NodeExecutionIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: nodeExecutionModel.Project,
				Domain:  nodeExecutionModel.Domain,
				Name:    nodeExecutionModel.Name,
			},
			NodeId: nodeExecutionModel.NodeID,
		}, transformers.ListExecutionTransformerOptions)
		if err != nil {
			return nil, err
		}
		nodeExecutions[idx] = nodeExecution
	}
	return nodeExecutions, nil
}

func (m *NodeExecutionManager) GetNodeExecution(
	ctx context.Context, request *admin.NodeExecutionGetRequest) (*admin.NodeExecution, error) {
	if err := validation.ValidateNodeExecutionIdentifier(request.GetId()); err != nil {
		logger.Debugf(ctx, "get node execution called with invalid identifier [%+v]: %v", request.GetId(), err)
	}
	ctx = getNodeExecutionContext(ctx, request.GetId())
	nodeExecutionModel, err := util.GetNodeExecutionModel(ctx, m.db, request.GetId())
	if err != nil {
		logger.Debugf(ctx, "Failed to get node execution with id [%+v] with err %v",
			request.GetId(), err)
		return nil, err
	}
	nodeExecution, err := m.transformNodeExecutionModel(ctx, *nodeExecutionModel, request.GetId(), nil)
	if err != nil {
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

	sortParameter, err := common.NewSortParameter(sortBy, models.NodeExecutionColumns)
	if err != nil {
		return nil, err
	}

	offset, err := validation.ValidateToken(requestToken)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"invalid pagination token %s for ListNodeExecutions", requestToken)
	}
	joinTableEntities := make(map[common.Entity]bool)
	for _, filter := range filters {
		joinTableEntities[filter.GetEntity()] = true
	}
	listInput := repoInterfaces.ListResourceInput{
		Limit:             int(limit),
		Offset:            offset,
		InlineFilters:     filters,
		SortParameter:     sortParameter,
		JoinTableEntities: joinTableEntities,
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
	nodeExecutionList, err := m.transformNodeExecutionModelList(ctx, output.NodeExecutions)
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
	ctx context.Context, request *admin.NodeExecutionListRequest) (*admin.NodeExecutionList, error) {
	// Check required fields
	if err := validation.ValidateNodeExecutionListRequest(request); err != nil {
		return nil, err
	}
	ctx = getExecutionContext(ctx, request.GetWorkflowExecutionId())

	identifierFilters, err := util.GetWorkflowExecutionIdentifierFilters(ctx, request.GetWorkflowExecutionId(), common.NodeExecution)
	if err != nil {
		return nil, err
	}
	var mapFilters []common.MapFilter
	if request.GetUniqueParentId() != "" {
		parentNodeExecution, err := util.GetNodeExecutionModel(ctx, m.db, &core.NodeExecutionIdentifier{
			ExecutionId: request.GetWorkflowExecutionId(),
			NodeId:      request.GetUniqueParentId(),
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
		ctx, identifierFilters, request.GetFilters(), request.GetLimit(), request.GetToken(), request.GetSortBy(), mapFilters)
}

// Filters on node executions matching the execution parameters (execution project, domain, and name) as well as the
// parent task execution id corresponding to the task execution identified in the request params.
func (m *NodeExecutionManager) ListNodeExecutionsForTask(
	ctx context.Context, request *admin.NodeExecutionForTaskListRequest) (*admin.NodeExecutionList, error) {
	// Check required fields
	if err := validation.ValidateNodeExecutionForTaskListRequest(request); err != nil {
		return nil, err
	}
	ctx = getTaskExecutionContext(ctx, request.GetTaskExecutionId())
	identifierFilters, err := util.GetWorkflowExecutionIdentifierFilters(
		ctx, request.GetTaskExecutionId().GetNodeExecutionId().GetExecutionId(), common.NodeExecution)
	if err != nil {
		return nil, err
	}
	parentTaskExecutionModel, err := util.GetTaskExecutionModel(ctx, m.db, request.GetTaskExecutionId())
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
		ctx, identifierFilters, request.GetFilters(), request.GetLimit(), request.GetToken(), request.GetSortBy(), nil)
}

func (m *NodeExecutionManager) GetNodeExecutionData(
	ctx context.Context, request *admin.NodeExecutionGetDataRequest) (*admin.NodeExecutionGetDataResponse, error) {
	if err := validation.ValidateNodeExecutionIdentifier(request.GetId()); err != nil {
		logger.Debugf(ctx, "can't get node execution data with invalid identifier [%+v]: %v", request.GetId(), err)
	}

	ctx = getNodeExecutionContext(ctx, request.GetId())
	nodeExecutionModel, err := util.GetNodeExecutionModel(ctx, m.db, request.GetId())
	if err != nil {
		logger.Debugf(ctx, "Failed to get node execution with id [%+v] with err %v",
			request.GetId(), err)
		return nil, err
	}

	nodeExecution, err := transformers.FromNodeExecutionModel(*nodeExecutionModel, transformers.DefaultExecutionTransformerOptions)
	if err != nil {
		logger.Debugf(ctx, "failed to transform node execution model [%+v] when fetching data: %v", request.GetId(), err)
		return nil, err
	}

	var inputs *core.LiteralMap
	var inputURLBlob *admin.UrlBlob
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		var err error
		inputs, inputURLBlob, err = util.GetInputs(groupCtx, m.urlData, m.config.ApplicationConfiguration().GetRemoteDataConfig(),
			m.storageClient, nodeExecution.GetInputUri())
		return err
	})

	var outputs *core.LiteralMap
	var outputURLBlob *admin.UrlBlob
	group.Go(func() error {
		var err error
		outputs, outputURLBlob, err = util.GetOutputs(groupCtx, m.urlData, m.config.ApplicationConfiguration().GetRemoteDataConfig(),
			m.storageClient, nodeExecution.GetClosure())
		return err
	})

	// Get the output variable map from workflow model TypedInterface
	var inputVariableMap, outputVariableMap *core.VariableMap
	group.Go(func() error {
		var err error

		switch nodeExecution.GetClosure().GetTargetMetadata().(type) {
		case *admin.NodeExecutionClosure_WorkflowNodeMetadata:
			execID := nodeExecution.GetClosure().GetTargetMetadata().(*admin.NodeExecutionClosure_WorkflowNodeMetadata).WorkflowNodeMetadata.GetExecutionId()
			workflowModel, err := m.db.WorkflowRepo().Get(groupCtx, repoInterfaces.Identifier{
				Project: execID.GetProject(),
				Domain:  execID.GetDomain(),
				Name:    execID.GetName(),
			})

			if err != nil {
				logger.Debugf(groupCtx, "Failed to get workflow model for node execution [%+v] with err %v", request.GetId(), err)
				return err
			}
			workflow, err := transformers.FromWorkflowModel(workflowModel)
			if err != nil {
				logger.Debugf(groupCtx, "Failed to transform workflow model for node execution [%+v] with err %v", request.GetId(), err)
				return err
			}

			inputVariableMap = workflow.GetClosure().GetCompiledWorkflow().GetPrimary().GetTemplate().GetInterface().GetInputs()
			outputVariableMap = workflow.GetClosure().GetCompiledWorkflow().GetPrimary().GetTemplate().GetInterface().GetOutputs()

		case *admin.NodeExecutionClosure_TaskNodeMetadata:
			execID := nodeExecution.GetId().GetExecutionId()
			executionModel, err := m.db.ExecutionRepo().Get(groupCtx, repoInterfaces.Identifier{
				Project: execID.GetProject(),
				Domain:  execID.GetDomain(),
				Name:    execID.GetName(),
			})

			if err != nil {
				logger.Debugf(groupCtx, "Failed to get execution model for node execution [%+v] with err %v", request.GetId(), err)
				return err
			}

			execution, err := transformers.FromExecutionModel(groupCtx, executionModel, transformers.DefaultExecutionTransformerOptions)

			if err != nil {
				logger.Debugf(groupCtx, "Failed to transform execution model for node execution [%+v] with err %v", request.GetId(), err)
				return err
			}

			taskModel, err := m.db.TaskRepo().Get(groupCtx, repoInterfaces.Identifier{
				Project: execID.GetProject(),
				Domain:  execID.GetDomain(),
				Name:    execution.GetSpec().GetLaunchPlan().GetName(),
				Version: execution.GetSpec().GetLaunchPlan().GetVersion(),
			})

			if err != nil {
				logger.Debugf(groupCtx, "Failed to get task model for node execution [%+v] with err %v", request.GetId(), err)
				return err
			}

			task, err := transformers.FromTaskModel(taskModel)

			if err != nil {
				logger.Debugf(groupCtx, "Failed to transform task model for node execution [%+v] with err %v", request.GetId(), err)
				return err
			}

			inputVariableMap = task.GetClosure().GetCompiledTask().GetTemplate().GetInterface().GetInputs()
			outputVariableMap = task.GetClosure().GetCompiledTask().GetTemplate().GetInterface().GetOutputs()
		}
		return err
	})

	err = group.Wait()
	if err != nil {
		return nil, err
	}

	response := &admin.NodeExecutionGetDataResponse{
		Inputs:            inputURLBlob,
		Outputs:           outputURLBlob,
		FullInputs:        inputs,
		FullOutputs:       outputs,
		FlyteUrls:         common.FlyteURLsFromNodeExecutionID(request.GetId(), nodeExecution.GetClosure() != nil && nodeExecution.GetClosure().GetDeckUri() != ""),
		InputVariableMap:  inputVariableMap,
		OutputVariableMap: outputVariableMap,
	}

	if len(nodeExecutionModel.DynamicWorkflowRemoteClosureReference) > 0 {
		closure, err := m.fetchDynamicWorkflowClosure(ctx, nodeExecutionModel.DynamicWorkflowRemoteClosureReference)
		if err != nil {
			return nil, err
		}

		response.DynamicWorkflow = &admin.DynamicWorkflowNodeMetadata{
			Id:                closure.GetPrimary().GetTemplate().GetId(),
			CompiledWorkflow:  closure,
			DynamicJobSpecUri: nodeExecution.GetClosure().GetDynamicJobSpecUri(),
		}
	}

	m.metrics.NodeExecutionInputBytes.Observe(float64(response.GetInputs().GetBytes()))
	if response.GetOutputs().GetBytes() > 0 {
		m.metrics.NodeExecutionOutputBytes.Observe(float64(response.GetOutputs().GetBytes()))
	} else if response.GetFullOutputs() != nil {
		m.metrics.NodeExecutionOutputBytes.Observe(float64(proto.Size(response.GetFullOutputs())))
	}

	return response, nil
}

func (m *NodeExecutionManager) fetchDynamicWorkflowClosure(ctx context.Context, ref string) (*core.CompiledWorkflowClosure, error) {
	closure := &core.CompiledWorkflowClosure{}
	err := m.storageClient.ReadProtobuf(ctx, storage.DataReference(ref), closure)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "Unable to read WorkflowClosure from location %s : %v", ref, err)
	}

	if wf := closure.GetPrimary(); wf == nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "Empty primary workflow definition in loaded dynamic workflow model.")
	} else if template := wf.GetTemplate(); template == nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "Empty primary workflow template in loaded dynamic workflow model.")
	}
	return closure, nil
}

func NewNodeExecutionManager(db repoInterfaces.Repository, config runtimeInterfaces.Configuration,
	storagePrefix []string, storageClient *storage.DataStore, scope promutils.Scope, urlData dataInterfaces.RemoteURLInterface,
	eventPublisher notificationInterfaces.Publisher, cloudEventPublisher cloudeventInterfaces.Publisher,
	eventWriter eventWriter.NodeExecutionEventWriter) interfaces.NodeExecutionInterface {
	metrics := nodeExecutionMetrics{
		Scope: scope,
		ActiveNodeExecutions: scope.MustNewGauge("active_node_executions",
			"overall count of active node executions"),
		NodeExecutionsCreated: scope.MustNewCounter("node_executions_created",
			"overall count of node executions created"),
		NodeExecutionsTerminated: labeled.NewCounter("node_executions_terminated",
			"overall count of terminated node executions", scope),
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
		db:                  db,
		config:              config,
		storagePrefix:       storagePrefix,
		storageClient:       storageClient,
		metrics:             metrics,
		urlData:             urlData,
		eventPublisher:      eventPublisher,
		dbEventWriter:       eventWriter,
		cloudEventPublisher: cloudEventPublisher,
	}
}
