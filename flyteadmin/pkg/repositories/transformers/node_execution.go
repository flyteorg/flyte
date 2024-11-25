package transformers

import (
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	genModel "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/gen/models"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

type ToNodeExecutionModelInput struct {
	Request                      *admin.NodeExecutionEventRequest
	ParentTaskExecutionID        *uint
	ParentID                     *uint
	DynamicWorkflowRemoteClosure string
	InlineEventDataPolicy        interfaces.InlineEventDataPolicy
	StorageClient                *storage.DataStore
}

func addNodeRunningState(request *admin.NodeExecutionEventRequest, nodeExecutionModel *models.NodeExecution,
	closure *admin.NodeExecutionClosure) error {
	occurredAt, err := ptypes.Timestamp(request.GetEvent().GetOccurredAt())
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal, "failed to unmarshal occurredAt with error: %v", err)
	}

	nodeExecutionModel.StartedAt = &occurredAt
	startedAtProto, err := ptypes.TimestampProto(occurredAt)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal,
			"failed to marshal occurredAt into a timestamp proto with error: %v", err)
	}
	closure.StartedAt = startedAtProto
	closure.DeckUri = request.Event.DeckUri
	return nil
}

func addTerminalState(
	ctx context.Context,
	request *admin.NodeExecutionEventRequest, nodeExecutionModel *models.NodeExecution,
	closure *admin.NodeExecutionClosure, inlineEventDataPolicy interfaces.InlineEventDataPolicy, storageClient *storage.DataStore) error {
	if closure.GetStartedAt() == nil {
		logger.Warning(context.Background(), "node execution is missing StartedAt")
	} else {
		endTime, err := ptypes.Timestamp(request.GetEvent().GetOccurredAt())
		if err != nil {
			return errors.NewFlyteAdminErrorf(
				codes.Internal, "Failed to parse node execution occurred at timestamp: %v", err)
		}
		nodeExecutionModel.Duration = endTime.Sub(*nodeExecutionModel.StartedAt)
		closure.Duration = ptypes.DurationProto(nodeExecutionModel.Duration)
	}

	// Serialize output results (if they exist)
	if request.GetEvent().GetOutputUri() != "" {
		closure.OutputResult = &admin.NodeExecutionClosure_OutputUri{
			OutputUri: request.GetEvent().GetOutputUri(),
		}
	} else if request.GetEvent().GetOutputData() != nil {
		switch inlineEventDataPolicy {
		case interfaces.InlineEventDataPolicyStoreInline:
			closure.OutputResult = &admin.NodeExecutionClosure_OutputData{
				OutputData: request.GetEvent().GetOutputData(),
			}
		default:
			logger.Debugf(ctx, "Offloading outputs per InlineEventDataPolicy")
			uri, err := common.OffloadLiteralMap(ctx, storageClient, request.GetEvent().GetOutputData(),
				request.GetEvent().GetId().GetExecutionId().GetProject(), request.GetEvent().GetId().GetExecutionId().GetDomain(), request.GetEvent().GetId().GetExecutionId().GetName(),
				request.GetEvent().GetId().GetNodeId(), OutputsObjectSuffix)
			if err != nil {
				return err
			}
			closure.OutputResult = &admin.NodeExecutionClosure_OutputUri{
				OutputUri: uri.String(),
			}
		}
	} else if request.GetEvent().GetError() != nil {
		closure.OutputResult = &admin.NodeExecutionClosure_Error{
			Error: request.GetEvent().GetError(),
		}
		k := request.GetEvent().GetError().GetKind().String()
		nodeExecutionModel.ErrorKind = &k
		nodeExecutionModel.ErrorCode = &request.Event.GetError().Code
	}
	closure.DeckUri = request.GetEvent().GetDeckUri()

	return nil
}

func CreateNodeExecutionModel(ctx context.Context, input ToNodeExecutionModelInput) (*models.NodeExecution, error) {
	nodeExecution := &models.NodeExecution{
		NodeExecutionKey: models.NodeExecutionKey{
			NodeID: input.Request.GetEvent().GetId().GetNodeId(),
			ExecutionKey: models.ExecutionKey{
				Project: input.Request.GetEvent().GetId().GetExecutionId().GetProject(),
				Domain:  input.Request.GetEvent().GetId().GetExecutionId().GetDomain(),
				Name:    input.Request.GetEvent().GetId().GetExecutionId().GetName(),
			},
		},
		Phase: input.Request.GetEvent().GetPhase().String(),
	}

	reportedAt := input.Request.GetEvent().GetReportedAt()
	if reportedAt == nil || (reportedAt.GetSeconds() == 0 && reportedAt.GetNanos() == 0) {
		reportedAt = input.Request.GetEvent().GetOccurredAt()
	}

	closure := admin.NodeExecutionClosure{
		Phase:     input.Request.GetEvent().GetPhase(),
		CreatedAt: input.Request.GetEvent().GetOccurredAt(),
		UpdatedAt: reportedAt,
	}

	nodeExecutionMetadata := admin.NodeExecutionMetaData{
		RetryGroup:   input.Request.GetEvent().GetRetryGroup(),
		SpecNodeId:   input.Request.GetEvent().GetSpecNodeId(),
		IsParentNode: input.Request.GetEvent().GetIsParent(),
		IsDynamic:    input.Request.GetEvent().GetIsDynamic(),
		IsArray:      input.Request.GetEvent().GetIsArray(),
	}
	err := handleNodeExecutionInputs(ctx, nodeExecution, input.Request, input.StorageClient)
	if err != nil {
		return nil, err
	}

	if input.Request.GetEvent().GetPhase() == core.NodeExecution_RUNNING {
		err := addNodeRunningState(input.Request, nodeExecution, &closure)
		if err != nil {
			return nil, err
		}
	}

	if common.IsNodeExecutionTerminal(input.Request.GetEvent().GetPhase()) {
		err := addTerminalState(ctx, input.Request, nodeExecution, &closure, input.InlineEventDataPolicy, input.StorageClient)
		if err != nil {
			return nil, err
		}
	}

	// Update TaskNodeMetadata, which includes caching information today.
	if input.Request.GetEvent().GetTaskNodeMetadata() != nil {
		targetMetadata := &admin.NodeExecutionClosure_TaskNodeMetadata{
			TaskNodeMetadata: &admin.TaskNodeMetadata{
				CheckpointUri: input.Request.GetEvent().GetTaskNodeMetadata().GetCheckpointUri(),
			},
		}
		if input.Request.GetEvent().GetTaskNodeMetadata().GetCatalogKey() != nil {
			st := input.Request.GetEvent().GetTaskNodeMetadata().GetCacheStatus().String()
			targetMetadata.TaskNodeMetadata.CacheStatus = input.Request.GetEvent().GetTaskNodeMetadata().GetCacheStatus()
			targetMetadata.TaskNodeMetadata.CatalogKey = input.Request.GetEvent().GetTaskNodeMetadata().GetCatalogKey()
			nodeExecution.CacheStatus = &st
		}
		closure.TargetMetadata = targetMetadata
	}

	marshaledClosure, err := proto.Marshal(&closure)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(
			codes.Internal, "failed to marshal node execution closure with error: %v", err)
	}
	marshaledNodeExecutionMetadata, err := proto.Marshal(&nodeExecutionMetadata)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(
			codes.Internal, "failed to marshal node execution metadata with error: %v", err)
	}
	nodeExecution.Closure = marshaledClosure
	nodeExecution.NodeExecutionMetadata = marshaledNodeExecutionMetadata
	nodeExecutionCreatedAt, err := ptypes.Timestamp(input.Request.GetEvent().GetOccurredAt())
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to read event timestamp")
	}
	nodeExecution.NodeExecutionCreatedAt = &nodeExecutionCreatedAt
	nodeExecutionUpdatedAt, err := ptypes.Timestamp(reportedAt)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to read event reported_at timestamp")
	}
	nodeExecution.NodeExecutionUpdatedAt = &nodeExecutionUpdatedAt
	if input.Request.GetEvent().GetParentTaskMetadata() != nil {
		nodeExecution.ParentTaskExecutionID = input.ParentTaskExecutionID
	}
	nodeExecution.ParentID = input.ParentID
	nodeExecution.DynamicWorkflowRemoteClosureReference = input.DynamicWorkflowRemoteClosure

	internalData := &genModel.NodeExecutionInternalData{
		EventVersion: input.Request.GetEvent().GetEventVersion(),
	}
	internalDataBytes, err := proto.Marshal(internalData)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to marshal node execution data with err: %v", err)
	}
	nodeExecution.InternalData = internalDataBytes
	return nodeExecution, nil
}

func UpdateNodeExecutionModel(
	ctx context.Context, request *admin.NodeExecutionEventRequest, nodeExecutionModel *models.NodeExecution,
	targetExecution *core.WorkflowExecutionIdentifier, dynamicWorkflowRemoteClosure string,
	inlineEventDataPolicy interfaces.InlineEventDataPolicy, storageClient *storage.DataStore) error {
	err := handleNodeExecutionInputs(ctx, nodeExecutionModel, request, storageClient)
	if err != nil {
		return err
	}
	var nodeExecutionClosure admin.NodeExecutionClosure
	err = proto.Unmarshal(nodeExecutionModel.Closure, &nodeExecutionClosure)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal,
			"failed to unmarshal node execution closure with error: %+v", err)
	}
	nodeExecutionModel.Phase = request.GetEvent().GetPhase().String()
	nodeExecutionClosure.Phase = request.GetEvent().GetPhase()
	reportedAt := request.GetEvent().GetReportedAt()
	if reportedAt == nil || (reportedAt.GetSeconds() == 0 && reportedAt.GetNanos() == 0) {
		reportedAt = request.GetEvent().GetOccurredAt()
	}
	nodeExecutionClosure.UpdatedAt = reportedAt

	if request.GetEvent().GetPhase() == core.NodeExecution_RUNNING {
		err := addNodeRunningState(request, nodeExecutionModel, &nodeExecutionClosure)
		if err != nil {
			return err
		}
	}
	if common.IsNodeExecutionTerminal(request.GetEvent().GetPhase()) {
		err := addTerminalState(ctx, request, nodeExecutionModel, &nodeExecutionClosure, inlineEventDataPolicy, storageClient)
		if err != nil {
			return err
		}
	}

	// If the node execution kicked off a workflow execution update the closure if it wasn't set
	if targetExecution != nil && nodeExecutionClosure.GetWorkflowNodeMetadata() == nil {
		nodeExecutionClosure.TargetMetadata = &admin.NodeExecutionClosure_WorkflowNodeMetadata{
			WorkflowNodeMetadata: &admin.WorkflowNodeMetadata{
				ExecutionId: targetExecution,
			},
		}
	}

	// Update TaskNodeMetadata, which includes caching information today.
	if request.GetEvent().GetTaskNodeMetadata() != nil {
		targetMetadata := &admin.NodeExecutionClosure_TaskNodeMetadata{
			TaskNodeMetadata: &admin.TaskNodeMetadata{
				CheckpointUri: request.GetEvent().GetTaskNodeMetadata().GetCheckpointUri(),
			},
		}
		if request.GetEvent().GetTaskNodeMetadata().GetCatalogKey() != nil {
			st := request.GetEvent().GetTaskNodeMetadata().GetCacheStatus().String()
			targetMetadata.TaskNodeMetadata.CacheStatus = request.GetEvent().GetTaskNodeMetadata().GetCacheStatus()
			targetMetadata.TaskNodeMetadata.CatalogKey = request.GetEvent().GetTaskNodeMetadata().GetCatalogKey()
			nodeExecutionModel.CacheStatus = &st
		}
		nodeExecutionClosure.TargetMetadata = targetMetadata

		// if this is a dynamic task then maintain the DynamicJobSpecUri
		dynamicWorkflowMetadata := request.GetEvent().GetTaskNodeMetadata().GetDynamicWorkflow()
		if dynamicWorkflowMetadata != nil && len(dynamicWorkflowMetadata.GetDynamicJobSpecUri()) > 0 {
			nodeExecutionClosure.DynamicJobSpecUri = dynamicWorkflowMetadata.GetDynamicJobSpecUri()
		}
	}

	marshaledClosure, err := proto.Marshal(&nodeExecutionClosure)
	if err != nil {
		return errors.NewFlyteAdminErrorf(
			codes.Internal, "failed to marshal node execution closure with error: %v", err)
	}

	nodeExecutionModel.Closure = marshaledClosure
	updatedAt, err := ptypes.Timestamp(reportedAt)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal, "failed to parse updated at timestamp")
	}
	nodeExecutionModel.NodeExecutionUpdatedAt = &updatedAt
	nodeExecutionModel.DynamicWorkflowRemoteClosureReference = dynamicWorkflowRemoteClosure

	// In the case of dynamic nodes reporting DYNAMIC_RUNNING, the IsParent and IsDynamic bits will be set for this event.
	// Update the node execution metadata accordingly.
	if request.GetEvent().GetIsParent() || request.GetEvent().GetIsDynamic() || request.GetEvent().GetIsArray() {
		var nodeExecutionMetadata admin.NodeExecutionMetaData
		if len(nodeExecutionModel.NodeExecutionMetadata) > 0 {
			if err := proto.Unmarshal(nodeExecutionModel.NodeExecutionMetadata, &nodeExecutionMetadata); err != nil {
				return errors.NewFlyteAdminErrorf(codes.Internal,
					"failed to unmarshal node execution metadata with error: %+v", err)
			}
		}
		// Not every event sends IsParent and IsDynamic as an artifact of how propeller handles dynamic nodes.
		// Only explicitly set the fields, when they're set in the event itself.
		if request.GetEvent().GetIsParent() {
			nodeExecutionMetadata.IsParentNode = true
		}
		if request.GetEvent().GetIsDynamic() {
			nodeExecutionMetadata.IsDynamic = true
		}
		if request.GetEvent().GetIsArray() {
			nodeExecutionMetadata.IsArray = true
		}
		nodeExecMetadataBytes, err := proto.Marshal(&nodeExecutionMetadata)
		if err != nil {
			return errors.NewFlyteAdminErrorf(codes.Internal,
				"failed to marshal node execution metadata with error: %+v", err)
		}
		nodeExecutionModel.NodeExecutionMetadata = nodeExecMetadataBytes
	}

	return nil
}

func FromNodeExecutionModel(nodeExecutionModel models.NodeExecution, opts *ExecutionTransformerOptions) (*admin.NodeExecution, error) {
	var closure admin.NodeExecutionClosure
	err := proto.Unmarshal(nodeExecutionModel.Closure, &closure)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to unmarshal closure")
	}

	if closure.GetError() != nil && opts != nil && opts.TrimErrorMessage && len(closure.GetError().GetMessage()) > 0 {
		trimmedErrOutputResult := closure.GetError()
		trimmedErrMessage := TrimErrorMessage(trimmedErrOutputResult.GetMessage())
		trimmedErrOutputResult.Message = trimmedErrMessage
		closure.OutputResult = &admin.NodeExecutionClosure_Error{
			Error: trimmedErrOutputResult,
		}
	}

	var nodeExecutionMetadata admin.NodeExecutionMetaData
	err = proto.Unmarshal(nodeExecutionModel.NodeExecutionMetadata, &nodeExecutionMetadata)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to unmarshal nodeExecutionMetadata")
	}
	// TODO: delete this block and references to preloading child node executions no earlier than Q3 2022
	// This is required for historical reasons because propeller did not always send IsParent or IsDynamic in events.
	if !(nodeExecutionMetadata.GetIsParentNode() || nodeExecutionMetadata.GetIsDynamic()) {
		if len(nodeExecutionModel.ChildNodeExecutions) > 0 {
			nodeExecutionMetadata.IsParentNode = true
			if len(nodeExecutionModel.DynamicWorkflowRemoteClosureReference) > 0 {
				nodeExecutionMetadata.IsDynamic = true
			}
		}
	}

	return &admin.NodeExecution{
		Id: &core.NodeExecutionIdentifier{
			NodeId: nodeExecutionModel.NodeID,
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: nodeExecutionModel.NodeExecutionKey.ExecutionKey.Project,
				Domain:  nodeExecutionModel.NodeExecutionKey.ExecutionKey.Domain,
				Name:    nodeExecutionModel.NodeExecutionKey.ExecutionKey.Name,
			},
		},
		InputUri: nodeExecutionModel.InputURI,
		Closure:  &closure,
		Metadata: &nodeExecutionMetadata,
	}, nil
}

func GetNodeExecutionInternalData(internalData []byte) (*genModel.NodeExecutionInternalData, error) {
	var nodeExecutionInternalData genModel.NodeExecutionInternalData
	if len(internalData) > 0 {
		err := proto.Unmarshal(internalData, &nodeExecutionInternalData)
		if err != nil {
			return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to unmarshal node execution data: %v", err)
		}
	}
	return &nodeExecutionInternalData, nil
}

func handleNodeExecutionInputs(ctx context.Context,
	nodeExecutionModel *models.NodeExecution,
	request *admin.NodeExecutionEventRequest,
	storageClient *storage.DataStore) error {
	if len(nodeExecutionModel.InputURI) > 0 {
		// Inputs are static over the duration of the node execution, no need to update them when they're already set
		return nil
	}
	switch request.GetEvent().GetInputValue().(type) {
	case *event.NodeExecutionEvent_InputUri:
		logger.Debugf(ctx, "saving node execution input URI [%s]", request.GetEvent().GetInputUri())
		nodeExecutionModel.InputURI = request.GetEvent().GetInputUri()
	case *event.NodeExecutionEvent_InputData:
		uri, err := common.OffloadLiteralMap(ctx, storageClient, request.GetEvent().GetInputData(),
			request.GetEvent().GetId().GetExecutionId().GetProject(), request.GetEvent().GetId().GetExecutionId().GetDomain(), request.GetEvent().GetId().GetExecutionId().GetName(),
			request.GetEvent().GetId().GetNodeId(), InputsObjectSuffix)
		if err != nil {
			return err
		}
		logger.Debugf(ctx, "offloaded node execution inputs to [%s]", uri)
		nodeExecutionModel.InputURI = uri.String()
	default:
		logger.Debugf(ctx, "request contained no input data")

	}
	return nil
}
