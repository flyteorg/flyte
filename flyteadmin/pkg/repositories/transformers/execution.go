package transformers

import (
	"context"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/lyft/flyteadmin/pkg/common"
	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/storage"
	"google.golang.org/grpc/codes"
)

// Request parameters for calls to CreateExecutionModel.
type CreateExecutionModelInput struct {
	WorkflowExecutionID   core.WorkflowExecutionIdentifier
	RequestSpec           *admin.ExecutionSpec
	LaunchPlanID          uint
	WorkflowID            uint
	Phase                 core.WorkflowExecution_Phase
	CreatedAt             time.Time
	Notifications         []*admin.Notification
	WorkflowIdentifier    *core.Identifier
	ParentNodeExecutionID uint
	Cluster               string
	InputsURI             storage.DataReference
	UserInputsURI         storage.DataReference
}

// Transforms a ExecutionCreateRequest to a Execution model
func CreateExecutionModel(input CreateExecutionModelInput) (*models.Execution, error) {
	spec, err := proto.Marshal(input.RequestSpec)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "Failed to serialize execution spec: %v", err)
	}
	createdAt, err := ptypes.TimestampProto(input.CreatedAt)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to serialize execution created at time")
	}
	closure := admin.ExecutionClosure{
		Phase:         input.Phase,
		CreatedAt:     createdAt,
		UpdatedAt:     createdAt,
		Notifications: input.Notifications,
		WorkflowId:    input.WorkflowIdentifier,
	}
	if input.Phase == core.WorkflowExecution_RUNNING {
		closure.StartedAt = createdAt
	}

	closureBytes, err := proto.Marshal(&closure)

	if err != nil {
		return nil, errors.NewFlyteAdminError(codes.Internal, "Failed to serialize launch plan status")
	}

	executionModel := &models.Execution{
		ExecutionKey: models.ExecutionKey{
			Project: input.WorkflowExecutionID.Project,
			Domain:  input.WorkflowExecutionID.Domain,
			Name:    input.WorkflowExecutionID.Name,
		},
		Spec:                  spec,
		Phase:                 input.Phase.String(),
		Closure:               closureBytes,
		LaunchPlanID:          input.LaunchPlanID,
		WorkflowID:            input.WorkflowID,
		ExecutionCreatedAt:    &input.CreatedAt,
		ExecutionUpdatedAt:    &input.CreatedAt,
		ParentNodeExecutionID: input.ParentNodeExecutionID,
		Cluster:               input.Cluster,
		InputsURI:             input.InputsURI,
		UserInputsURI:         input.UserInputsURI,
	}
	if input.RequestSpec.Metadata != nil {
		executionModel.Mode = int32(input.RequestSpec.Metadata.Mode)
	}

	return executionModel, nil
}

// Updates an existing model given a WorkflowExecution event.
func UpdateExecutionModelState(
	execution *models.Execution, request admin.WorkflowExecutionEventRequest, abortCause *string) error {
	var executionClosure admin.ExecutionClosure
	err := proto.Unmarshal(execution.Closure, &executionClosure)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal, "Failed to unmarshal execution closure: %v", err)
	}
	executionClosure.Phase = request.Event.Phase
	executionClosure.UpdatedAt = request.Event.OccurredAt
	execution.Phase = request.Event.Phase.String()

	occurredAtTimestamp, err := ptypes.Timestamp(request.Event.OccurredAt)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal, "Failed to parse OccurredAt: %v", err)
	}
	execution.ExecutionUpdatedAt = &occurredAtTimestamp

	// only mark the execution started when we get the initial running event
	if request.Event.Phase == core.WorkflowExecution_RUNNING {
		execution.StartedAt = &occurredAtTimestamp
		executionClosure.StartedAt = request.Event.OccurredAt
	} else if common.IsExecutionTerminal(request.Event.Phase) {
		if execution.StartedAt != nil {
			execution.Duration = occurredAtTimestamp.Sub(*execution.StartedAt)
			executionClosure.Duration = ptypes.DurationProto(execution.Duration)
		} else {
			logger.Infof(context.Background(),
				"Cannot compute duration because startedAt was never set, requestId: %v", request.RequestId)
		}
	}

	if request.Event.GetOutputUri() != "" {
		executionClosure.OutputResult = &admin.ExecutionClosure_Outputs{
			Outputs: &admin.LiteralMapBlob{
				Data: &admin.LiteralMapBlob_Uri{
					Uri: request.Event.GetOutputUri(),
				},
			},
		}
	}

	if request.Event.GetError() != nil {
		executionClosure.OutputResult = &admin.ExecutionClosure_Error{
			Error: request.Event.GetError(),
		}
	}
	marshaledClosure, err := proto.Marshal(&executionClosure)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal, "Failed to marshal execution closure: %v", err)
	}
	execution.Closure = marshaledClosure
	if abortCause != nil {
		execution.AbortCause = *abortCause
	}
	return nil
}

func GetExecutionIdentifier(executionModel *models.Execution) core.WorkflowExecutionIdentifier {
	return core.WorkflowExecutionIdentifier{
		Project: executionModel.Project,
		Domain:  executionModel.Domain,
		Name:    executionModel.Name,
	}
}

func FromExecutionModel(executionModel models.Execution) (*admin.Execution, error) {
	var spec admin.ExecutionSpec
	err := proto.Unmarshal(executionModel.Spec, &spec)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to unmarshal spec")
	}
	var closure admin.ExecutionClosure
	err = proto.Unmarshal(executionModel.Closure, &closure)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to unmarshal closure")
	}
	id := GetExecutionIdentifier(&executionModel)
	if executionModel.Phase == core.WorkflowExecution_ABORTED.String() {
		closure.OutputResult = &admin.ExecutionClosure_AbortCause{
			AbortCause: executionModel.AbortCause,
		}
	}

	// TODO: Clear deprecated fields to reduce message size.
	// spec.Inputs = nil
	// closure.ComputedInputs = nil
	return &admin.Execution{
		Id:      &id,
		Spec:    &spec,
		Closure: &closure,
	}, nil
}

func FromExecutionModelWithReferenceExecution(executionModel models.Execution, referenceExecutionID *core.WorkflowExecutionIdentifier) (
	*admin.Execution, error) {
	execution, err := FromExecutionModel(executionModel)
	if err != nil {
		return nil, err
	}
	if referenceExecutionID != nil && execution.Spec.Metadata != nil &&
		execution.Spec.Metadata.Mode == admin.ExecutionMetadata_RELAUNCH {
		execution.Spec.Metadata.ReferenceExecution = referenceExecutionID
	}
	return execution, nil
}

func FromExecutionModels(executionModels []models.Execution) ([]*admin.Execution, error) {
	executions := make([]*admin.Execution, len(executionModels))
	for idx, executionModel := range executionModels {
		execution, err := FromExecutionModel(executionModel)
		if err != nil {
			return nil, err
		}
		executions[idx] = execution
	}
	return executions, nil
}
