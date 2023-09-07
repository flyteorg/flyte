package transformers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/apimachinery/pkg/util/sets"
)

const trimmedErrMessageLen = 100

var clusterReassignablePhases = sets.NewString(core.WorkflowExecution_UNDEFINED.String(), core.WorkflowExecution_QUEUED.String())

// CreateExecutionModelInput encapsulates request parameters for calls to CreateExecutionModel.
type CreateExecutionModelInput struct {
	WorkflowExecutionID   core.WorkflowExecutionIdentifier
	RequestSpec           *admin.ExecutionSpec
	LaunchPlanID          uint
	WorkflowID            uint
	TaskID                uint
	Phase                 core.WorkflowExecution_Phase
	CreatedAt             time.Time
	Notifications         []*admin.Notification
	WorkflowIdentifier    *core.Identifier
	ParentNodeExecutionID uint
	SourceExecutionID     uint
	Cluster               string
	InputsURI             storage.DataReference
	UserInputsURI         storage.DataReference
	SecurityContext       *core.SecurityContext
	LaunchEntity          core.ResourceType
	Namespace             string
}

type ExecutionTransformerOptions struct {
	TrimErrorMessage bool
	DefaultNamespace string
}

var DefaultExecutionTransformerOptions = &ExecutionTransformerOptions{}
var ListExecutionTransformerOptions = &ExecutionTransformerOptions{
	TrimErrorMessage: true,
}

// CreateExecutionModel transforms a ExecutionCreateRequest to a Execution model
func CreateExecutionModel(input CreateExecutionModelInput) (*models.Execution, error) {
	requestSpec := input.RequestSpec
	if requestSpec.Metadata == nil {
		requestSpec.Metadata = &admin.ExecutionMetadata{}
	}
	requestSpec.Metadata.SystemMetadata = &admin.SystemMetadata{
		ExecutionCluster: input.Cluster,
		Namespace:        input.Namespace,
	}
	requestSpec.SecurityContext = input.SecurityContext
	spec, err := proto.Marshal(requestSpec)
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
		StateChangeDetails: &admin.ExecutionStateChangeDetails{
			State:      admin.ExecutionState_EXECUTION_ACTIVE,
			Principal:  requestSpec.Metadata.Principal,
			OccurredAt: createdAt,
		},
	}
	if input.Phase == core.WorkflowExecution_RUNNING {
		closure.StartedAt = createdAt
	}

	closureBytes, err := proto.Marshal(&closure)

	if err != nil {
		return nil, errors.NewFlyteAdminError(codes.Internal, "Failed to serialize launch plan status")
	}

	activeExecution := int32(admin.ExecutionState_EXECUTION_ACTIVE)
	tags := make([]models.AdminTag, len(input.RequestSpec.Tags))
	for i, tag := range input.RequestSpec.Tags {
		tags[i] = models.AdminTag{Name: tag}
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
		WorkflowID:            input.WorkflowID,
		ExecutionCreatedAt:    &input.CreatedAt,
		ExecutionUpdatedAt:    &input.CreatedAt,
		ParentNodeExecutionID: input.ParentNodeExecutionID,
		SourceExecutionID:     input.SourceExecutionID,
		Cluster:               input.Cluster,
		InputsURI:             input.InputsURI,
		UserInputsURI:         input.UserInputsURI,
		User:                  requestSpec.Metadata.Principal,
		State:                 &activeExecution,
		LaunchEntity:          strings.ToLower(input.LaunchEntity.String()),
		Tags:                  tags,
	}
	// A reference launch entity can be one of either or a task OR launch plan. Traditionally, workflows are executed
	// with a reference launch plan which is why this behavior is the default below.
	if input.TaskID > 0 {
		executionModel.TaskID = input.TaskID
	} else {
		executionModel.LaunchPlanID = input.LaunchPlanID
	}
	if input.RequestSpec.Metadata != nil {
		executionModel.Mode = int32(input.RequestSpec.Metadata.Mode)
	}

	return executionModel, nil
}

func reassignCluster(ctx context.Context, cluster string, executionID *core.WorkflowExecutionIdentifier, execution *models.Execution) error {
	logger.Debugf(ctx, "Updating cluster for execution [%v] with existing recorded cluster [%s] and setting to cluster [%s]",
		executionID, execution.Cluster, cluster)
	execution.Cluster = cluster
	var executionSpec admin.ExecutionSpec
	err := proto.Unmarshal(execution.Spec, &executionSpec)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal, "Failed to unmarshal execution spec: %v", err)
	}
	if executionSpec.Metadata == nil {
		executionSpec.Metadata = &admin.ExecutionMetadata{}
	}
	if executionSpec.Metadata.SystemMetadata == nil {
		executionSpec.Metadata.SystemMetadata = &admin.SystemMetadata{}
	}
	executionSpec.Metadata.SystemMetadata.ExecutionCluster = cluster
	marshaledSpec, err := proto.Marshal(&executionSpec)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal, "Failed to marshal execution spec: %v", err)
	}
	execution.Spec = marshaledSpec
	return nil
}

// Updates an existing model given a WorkflowExecution event.
func UpdateExecutionModelState(
	ctx context.Context,
	execution *models.Execution, request admin.WorkflowExecutionEventRequest,
	inlineEventDataPolicy interfaces.InlineEventDataPolicy, storageClient *storage.DataStore) error {
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

	// Default or empty cluster values do not require updating the execution model.
	ignoreClusterFromEvent := len(request.Event.ProducerId) == 0 || request.Event.ProducerId == common.DefaultProducerID
	logger.Debugf(ctx, "Producer Id [%v]. IgnoreClusterFromEvent [%v]", request.Event.ProducerId, ignoreClusterFromEvent)
	if !ignoreClusterFromEvent {
		if clusterReassignablePhases.Has(execution.Phase) {
			if err := reassignCluster(ctx, request.Event.ProducerId, request.Event.ExecutionId, execution); err != nil {
				return err
			}
		} else if execution.Cluster != request.Event.ProducerId {
			errorMsg := fmt.Sprintf("Cannot accept events for running/terminated execution [%v] from cluster [%s],"+
				"expected events to originate from [%s]",
				request.Event.ExecutionId, request.Event.ProducerId, execution.Cluster)
			return errors.NewIncompatibleClusterError(ctx, errorMsg, execution.Cluster)
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
	} else if request.Event.GetOutputData() != nil {
		switch inlineEventDataPolicy {
		case interfaces.InlineEventDataPolicyStoreInline:
			executionClosure.OutputResult = &admin.ExecutionClosure_OutputData{
				OutputData: request.Event.GetOutputData(),
			}
		default:
			logger.Debugf(ctx, "Offloading outputs per InlineEventDataPolicy")
			uri, err := common.OffloadLiteralMap(ctx, storageClient, request.Event.GetOutputData(),
				request.Event.ExecutionId.Project, request.Event.ExecutionId.Domain, request.Event.ExecutionId.Name, OutputsObjectSuffix)
			if err != nil {
				return err
			}
			executionClosure.OutputResult = &admin.ExecutionClosure_Outputs{
				Outputs: &admin.LiteralMapBlob{
					Data: &admin.LiteralMapBlob_Uri{
						Uri: uri.String(),
					},
				},
			}
		}
	} else if request.Event.GetError() != nil {
		executionClosure.OutputResult = &admin.ExecutionClosure_Error{
			Error: request.Event.GetError(),
		}
		k := request.Event.GetError().Kind.String()
		execution.ErrorKind = &k
		execution.ErrorCode = &request.Event.GetError().Code
	}
	marshaledClosure, err := proto.Marshal(&executionClosure)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal, "Failed to marshal execution closure: %v", err)
	}
	execution.Closure = marshaledClosure
	return nil
}

// UpdateExecutionModelStateChangeDetails Updates an existing model with stateUpdateTo, stateUpdateBy and
// statedUpdatedAt details from the request
func UpdateExecutionModelStateChangeDetails(executionModel *models.Execution, stateUpdatedTo admin.ExecutionState,
	stateUpdatedAt time.Time, stateUpdatedBy string) error {

	var closure admin.ExecutionClosure
	err := proto.Unmarshal(executionModel.Closure, &closure)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal, "Failed to unmarshal execution closure: %v", err)
	}
	// Update the indexed columns
	stateInt := int32(stateUpdatedTo)
	executionModel.State = &stateInt

	// Update the closure with the same
	var stateUpdatedAtProto *timestamppb.Timestamp
	// Default use the createdAt timestamp as the state change occurredAt time
	if stateUpdatedAtProto, err = ptypes.TimestampProto(stateUpdatedAt); err != nil {
		return err
	}
	closure.StateChangeDetails = &admin.ExecutionStateChangeDetails{
		State:      stateUpdatedTo,
		Principal:  stateUpdatedBy,
		OccurredAt: stateUpdatedAtProto,
	}
	marshaledClosure, err := proto.Marshal(&closure)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal, "Failed to marshal execution closure: %v", err)
	}
	executionModel.Closure = marshaledClosure
	return nil
}

// The execution abort metadata is recorded but the phase is not actually updated *until* the abort event is propagated
// by flytepropeller. The metadata is preemptively saved at the time of the abort.
func SetExecutionAborting(execution *models.Execution, cause, principal string) error {
	var closure admin.ExecutionClosure
	err := proto.Unmarshal(execution.Closure, &closure)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal, "Failed to unmarshal execution closure: %v", err)
	}
	closure.OutputResult = &admin.ExecutionClosure_AbortMetadata{
		AbortMetadata: &admin.AbortMetadata{
			Cause:     cause,
			Principal: principal,
		},
	}
	closure.Phase = core.WorkflowExecution_ABORTING
	marshaledClosure, err := proto.Marshal(&closure)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal, "Failed to marshal execution closure: %v", err)
	}
	execution.Closure = marshaledClosure
	execution.AbortCause = cause
	execution.Phase = core.WorkflowExecution_ABORTING.String()
	return nil
}

func GetExecutionIdentifier(executionModel *models.Execution) core.WorkflowExecutionIdentifier {
	return core.WorkflowExecutionIdentifier{
		Project: executionModel.Project,
		Domain:  executionModel.Domain,
		Name:    executionModel.Name,
	}
}

func FromExecutionModel(ctx context.Context, executionModel models.Execution, opts *ExecutionTransformerOptions) (*admin.Execution, error) {
	var spec admin.ExecutionSpec
	var err error
	if err = proto.Unmarshal(executionModel.Spec, &spec); err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to unmarshal spec")
	}
	if len(opts.DefaultNamespace) > 0 {
		if spec.Metadata == nil {
			spec.Metadata = &admin.ExecutionMetadata{}
		}
		if spec.Metadata.SystemMetadata == nil {
			spec.Metadata.SystemMetadata = &admin.SystemMetadata{}
		}
		if len(spec.GetMetadata().GetSystemMetadata().Namespace) == 0 {
			logger.Infof(ctx, "setting execution system metadata namespace to [%s]", opts.DefaultNamespace)
			spec.Metadata.SystemMetadata.Namespace = opts.DefaultNamespace
		}
	}

	var closure admin.ExecutionClosure
	if err = proto.Unmarshal(executionModel.Closure, &closure); err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to unmarshal closure")
	}
	if closure.GetError() != nil && opts != nil && opts.TrimErrorMessage && len(closure.GetError().Message) > 0 {
		trimmedErrOutputResult := closure.GetError()
		trimmedErrMessage := TrimErrorMessage(trimmedErrOutputResult.GetMessage())
		trimmedErrOutputResult.Message = trimmedErrMessage
		closure.OutputResult = &admin.ExecutionClosure_Error{
			Error: trimmedErrOutputResult,
		}
	}

	if closure.StateChangeDetails == nil {
		// Update execution state details from model for older executions
		if closure.StateChangeDetails, err = PopulateDefaultStateChangeDetails(executionModel); err != nil {
			return nil, err
		}
	}

	id := GetExecutionIdentifier(&executionModel)
	if executionModel.Phase == core.WorkflowExecution_ABORTED.String() && closure.GetAbortMetadata() == nil {
		// In the case of data predating the AbortMetadata field we manually set it in the closure only
		// if it does not yet exist.
		closure.OutputResult = &admin.ExecutionClosure_AbortMetadata{
			AbortMetadata: &admin.AbortMetadata{
				Cause: executionModel.AbortCause,
			},
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

// PopulateDefaultStateChangeDetails used to populate execution state change details for older executions which donot
// have these details captured. Hence we construct a default state change details from existing data model.
func PopulateDefaultStateChangeDetails(executionModel models.Execution) (*admin.ExecutionStateChangeDetails, error) {
	var err error
	var occurredAt *timestamppb.Timestamp

	// Default use the createdAt timestamp as the state change occurredAt time
	if occurredAt, err = ptypes.TimestampProto(executionModel.CreatedAt); err != nil {
		return nil, err
	}

	return &admin.ExecutionStateChangeDetails{
		State:      admin.ExecutionState_EXECUTION_ACTIVE,
		OccurredAt: occurredAt,
		Principal:  executionModel.User,
	}, nil
}

func FromExecutionModels(ctx context.Context, executionModels []models.Execution, opts *ExecutionTransformerOptions) ([]*admin.Execution, error) {
	executions := make([]*admin.Execution, len(executionModels))
	for idx, executionModel := range executionModels {
		execution, err := FromExecutionModel(ctx, executionModel, opts)
		if err != nil {
			return nil, err
		}
		executions[idx] = execution
	}
	return executions, nil
}

// TrimErrorMessage return the smallest possible trimmed error message >= trimmedErrMessageLen bytes in length that still forms a valid utf-8 string
func TrimErrorMessage(errMsg string) string {
	if len(errMsg) <= trimmedErrMessageLen {
		return errMsg
	}
	return strings.ToValidUTF8(errMsg[:trimmedErrMessageLen], "")
}
