package transformers

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"

	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/storage"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteadmin/pkg/common"

	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/golang/protobuf/ptypes"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	genModel "github.com/flyteorg/flyteadmin/pkg/repositories/gen/models"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"google.golang.org/grpc/codes"
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
	occurredAt, err := ptypes.Timestamp(request.Event.OccurredAt)
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
	return nil
}

func addTerminalState(
	ctx context.Context,
	request *admin.NodeExecutionEventRequest, nodeExecutionModel *models.NodeExecution,
	closure *admin.NodeExecutionClosure, inlineEventDataPolicy interfaces.InlineEventDataPolicy, storageClient *storage.DataStore) error {
	if closure.StartedAt == nil {
		logger.Warning(context.Background(), "node execution is missing StartedAt")
	} else {
		endTime, err := ptypes.Timestamp(request.Event.OccurredAt)
		if err != nil {
			return errors.NewFlyteAdminErrorf(
				codes.Internal, "Failed to parse node execution occurred at timestamp: %v", err)
		}
		nodeExecutionModel.Duration = endTime.Sub(*nodeExecutionModel.StartedAt)
		closure.Duration = ptypes.DurationProto(nodeExecutionModel.Duration)
	}

	// Serialize output results (if they exist)
	if request.Event.GetOutputUri() != "" {
		closure.OutputResult = &admin.NodeExecutionClosure_OutputUri{
			OutputUri: request.Event.GetOutputUri(),
		}
	} else if request.Event.GetOutputData() != nil {
		switch inlineEventDataPolicy {
		case interfaces.InlineEventDataPolicyStoreInline:
			closure.OutputResult = &admin.NodeExecutionClosure_OutputData{
				OutputData: request.Event.GetOutputData(),
			}
		default:
			logger.Debugf(ctx, "Offloading outputs per InlineEventDataPolicy")
			uri, err := common.OffloadLiteralMap(ctx, storageClient, request.Event.GetOutputData(),
				request.Event.Id.ExecutionId.Project, request.Event.Id.ExecutionId.Domain, request.Event.Id.ExecutionId.Name,
				request.Event.Id.NodeId, OutputsObjectSuffix)
			if err != nil {
				return err
			}
			closure.OutputResult = &admin.NodeExecutionClosure_OutputUri{
				OutputUri: uri.String(),
			}
		}
	} else if request.Event.GetError() != nil {
		closure.OutputResult = &admin.NodeExecutionClosure_Error{
			Error: request.Event.GetError(),
		}
		k := request.Event.GetError().Kind.String()
		nodeExecutionModel.ErrorKind = &k
		nodeExecutionModel.ErrorCode = &request.Event.GetError().Code
	}
	closure.DeckUri = request.Event.DeckUri

	return nil
}

func CreateNodeExecutionModel(ctx context.Context, input ToNodeExecutionModelInput) (*models.NodeExecution, error) {
	nodeExecution := &models.NodeExecution{
		NodeExecutionKey: models.NodeExecutionKey{
			NodeID: input.Request.Event.Id.NodeId,
			ExecutionKey: models.ExecutionKey{
				Project: input.Request.Event.Id.ExecutionId.Project,
				Domain:  input.Request.Event.Id.ExecutionId.Domain,
				Name:    input.Request.Event.Id.ExecutionId.Name,
			},
		},
		Phase: input.Request.Event.Phase.String(),
	}

	reportedAt := input.Request.Event.ReportedAt
	if reportedAt == nil || (reportedAt.Seconds == 0 && reportedAt.Nanos == 0) {
		reportedAt = input.Request.Event.OccurredAt
	}

	closure := admin.NodeExecutionClosure{
		Phase:     input.Request.Event.Phase,
		CreatedAt: input.Request.Event.OccurredAt,
		UpdatedAt: reportedAt,
	}

	nodeExecutionMetadata := admin.NodeExecutionMetaData{
		RetryGroup:   input.Request.Event.RetryGroup,
		SpecNodeId:   input.Request.Event.SpecNodeId,
		IsParentNode: input.Request.Event.IsParent,
		IsDynamic:    input.Request.Event.IsDynamic,
	}
	err := handleNodeExecutionInputs(ctx, nodeExecution, input.Request, input.StorageClient)
	if err != nil {
		return nil, err
	}

	if input.Request.Event.Phase == core.NodeExecution_RUNNING {
		err := addNodeRunningState(input.Request, nodeExecution, &closure)
		if err != nil {
			return nil, err
		}
	}

	if common.IsNodeExecutionTerminal(input.Request.Event.Phase) {
		err := addTerminalState(ctx, input.Request, nodeExecution, &closure, input.InlineEventDataPolicy, input.StorageClient)
		if err != nil {
			return nil, err
		}
	}

	// Update TaskNodeMetadata, which includes caching information today.
	if input.Request.Event.GetTaskNodeMetadata() != nil {
		targetMetadata := &admin.NodeExecutionClosure_TaskNodeMetadata{
			TaskNodeMetadata: &admin.TaskNodeMetadata{
				CheckpointUri: input.Request.Event.GetTaskNodeMetadata().CheckpointUri,
			},
		}
		if input.Request.Event.GetTaskNodeMetadata().CatalogKey != nil {
			st := input.Request.Event.GetTaskNodeMetadata().GetCacheStatus().String()
			targetMetadata.TaskNodeMetadata.CacheStatus = input.Request.Event.GetTaskNodeMetadata().GetCacheStatus()
			targetMetadata.TaskNodeMetadata.CatalogKey = input.Request.Event.GetTaskNodeMetadata().GetCatalogKey()
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
	nodeExecutionCreatedAt, err := ptypes.Timestamp(input.Request.Event.OccurredAt)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to read event timestamp")
	}
	nodeExecution.NodeExecutionCreatedAt = &nodeExecutionCreatedAt
	nodeExecutionUpdatedAt, err := ptypes.Timestamp(reportedAt)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to read event reported_at timestamp")
	}
	nodeExecution.NodeExecutionUpdatedAt = &nodeExecutionUpdatedAt
	if input.Request.Event.ParentTaskMetadata != nil {
		nodeExecution.ParentTaskExecutionID = input.ParentTaskExecutionID
	}
	nodeExecution.ParentID = input.ParentID
	nodeExecution.DynamicWorkflowRemoteClosureReference = input.DynamicWorkflowRemoteClosure

	internalData := &genModel.NodeExecutionInternalData{
		EventVersion: input.Request.Event.EventVersion,
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
	nodeExecutionModel.Phase = request.Event.Phase.String()
	nodeExecutionClosure.Phase = request.Event.Phase
	reportedAt := request.Event.ReportedAt
	if reportedAt == nil || (reportedAt.Seconds == 0 && reportedAt.Nanos == 0) {
		reportedAt = request.Event.OccurredAt
	}
	nodeExecutionClosure.UpdatedAt = reportedAt

	if request.Event.Phase == core.NodeExecution_RUNNING {
		err := addNodeRunningState(request, nodeExecutionModel, &nodeExecutionClosure)
		if err != nil {
			return err
		}
	}
	if common.IsNodeExecutionTerminal(request.Event.Phase) {
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
	if request.Event.GetTaskNodeMetadata() != nil {
		targetMetadata := &admin.NodeExecutionClosure_TaskNodeMetadata{
			TaskNodeMetadata: &admin.TaskNodeMetadata{
				CheckpointUri: request.Event.GetTaskNodeMetadata().CheckpointUri,
			},
		}
		if request.Event.GetTaskNodeMetadata().CatalogKey != nil {
			st := request.Event.GetTaskNodeMetadata().GetCacheStatus().String()
			targetMetadata.TaskNodeMetadata.CacheStatus = request.Event.GetTaskNodeMetadata().GetCacheStatus()
			targetMetadata.TaskNodeMetadata.CatalogKey = request.Event.GetTaskNodeMetadata().GetCatalogKey()
			nodeExecutionModel.CacheStatus = &st
		}
		nodeExecutionClosure.TargetMetadata = targetMetadata

		// if this is a dynamic task then maintain the DynamicJobSpecUri
		dynamicWorkflowMetadata := request.Event.GetTaskNodeMetadata().DynamicWorkflow
		if dynamicWorkflowMetadata != nil && len(dynamicWorkflowMetadata.DynamicJobSpecUri) > 0 {
			nodeExecutionClosure.DynamicJobSpecUri = dynamicWorkflowMetadata.DynamicJobSpecUri
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
	if request.Event.IsParent || request.Event.IsDynamic {
		var nodeExecutionMetadata admin.NodeExecutionMetaData
		if len(nodeExecutionModel.NodeExecutionMetadata) > 0 {
			if err := proto.Unmarshal(nodeExecutionModel.NodeExecutionMetadata, &nodeExecutionMetadata); err != nil {
				return errors.NewFlyteAdminErrorf(codes.Internal,
					"failed to unmarshal node execution metadata with error: %+v", err)
			}
		}
		// Not every event sends IsParent and IsDynamic as an artifact of how propeller handles dynamic nodes.
		// Only explicitly set the fields, when they're set in the event itself.
		if request.Event.IsParent {
			nodeExecutionMetadata.IsParentNode = true
		}
		if request.Event.IsDynamic {
			nodeExecutionMetadata.IsDynamic = true
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

	if closure.GetError() != nil && opts != nil && opts.TrimErrorMessage && len(closure.GetError().Message) > 0 {
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
	if !(nodeExecutionMetadata.IsParentNode || nodeExecutionMetadata.IsDynamic) {
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
	switch request.Event.GetInputValue().(type) {
	case *event.NodeExecutionEvent_InputUri:
		logger.Debugf(ctx, "saving node execution input URI [%s]", request.Event.GetInputUri())
		nodeExecutionModel.InputURI = request.Event.GetInputUri()
	case *event.NodeExecutionEvent_InputData:
		uri, err := common.OffloadLiteralMap(ctx, storageClient, request.Event.GetInputData(),
			request.Event.Id.ExecutionId.Project, request.Event.Id.ExecutionId.Domain, request.Event.Id.ExecutionId.Name,
			request.Event.Id.NodeId, InputsObjectSuffix)
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
