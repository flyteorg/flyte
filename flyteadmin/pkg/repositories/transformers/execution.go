package transformers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	flyteErrs "github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

const trimmedErrMessageLen = 100

var clusterReassignablePhases = sets.NewString(core.WorkflowExecution_UNDEFINED.String(), core.WorkflowExecution_QUEUED.String())

// CreateExecutionModelInput encapsulates request parameters for calls to CreateExecutionModel.
type CreateExecutionModelInput struct {
	WorkflowExecutionID   *core.WorkflowExecutionIdentifier
	RequestSpec           *admin.ExecutionSpec
	LaunchPlanID          uint
	WorkflowID            uint
	TaskID                uint
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
	Error                 error
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
	if requestSpec.GetMetadata() == nil {
		requestSpec.Metadata = &admin.ExecutionMetadata{}
	}
	requestSpec.Metadata.SystemMetadata = &admin.SystemMetadata{
		ExecutionCluster: input.Cluster,
		Namespace:        input.Namespace,
	}
	requestSpec.SecurityContext = input.SecurityContext
	spec, err := proto.Marshal(requestSpec)
	if err != nil {
		return nil, flyteErrs.NewFlyteAdminErrorf(codes.Internal, "Failed to serialize execution spec: %v", err)
	}
	createdAt := timestamppb.New(input.CreatedAt)
	closure := admin.ExecutionClosure{
		CreatedAt:     createdAt,
		UpdatedAt:     createdAt,
		Notifications: input.Notifications,
		WorkflowId:    input.WorkflowIdentifier,
		StateChangeDetails: &admin.ExecutionStateChangeDetails{
			State:      admin.ExecutionState_EXECUTION_ACTIVE,
			Principal:  requestSpec.GetMetadata().GetPrincipal(),
			OccurredAt: createdAt,
		},
	}
	if input.Error != nil {
		closure.Phase = core.WorkflowExecution_FAILED
		execErr := &core.ExecutionError{
			Code:    "Unknown",
			Message: input.Error.Error(),
			Kind:    core.ExecutionError_SYSTEM,
		}

		var adminErr flyteErrs.FlyteAdminError
		if errors.As(input.Error, &adminErr) {
			execErr.Code = adminErr.Code().String()
			execErr.Message = adminErr.Error()
			if adminErr.Code() == codes.InvalidArgument {
				execErr.Kind = core.ExecutionError_USER
			}
		}
		closure.OutputResult = &admin.ExecutionClosure_Error{Error: execErr}
	}

	closureBytes, err := proto.Marshal(&closure)

	if err != nil {
		return nil, flyteErrs.NewFlyteAdminError(codes.Internal, "Failed to serialize launch plan status")
	}

	activeExecution := int32(admin.ExecutionState_EXECUTION_ACTIVE)

	executionModel := &models.Execution{
		ExecutionKey: models.ExecutionKey{
			Project: input.WorkflowExecutionID.GetProject(),
			Domain:  input.WorkflowExecutionID.GetDomain(),
			Name:    input.WorkflowExecutionID.GetName(),
		},
		Spec:                  spec,
		Phase:                 closure.GetPhase().String(),
		Closure:               closureBytes,
		WorkflowID:            input.WorkflowID,
		ExecutionCreatedAt:    &input.CreatedAt,
		ExecutionUpdatedAt:    &input.CreatedAt,
		ParentNodeExecutionID: input.ParentNodeExecutionID,
		SourceExecutionID:     input.SourceExecutionID,
		Cluster:               input.Cluster,
		InputsURI:             input.InputsURI,
		UserInputsURI:         input.UserInputsURI,
		User:                  requestSpec.GetMetadata().GetPrincipal(),
		State:                 &activeExecution,
		LaunchEntity:          strings.ToLower(input.LaunchEntity.String()),
	}
	// A reference launch entity can be one of either or a task OR launch plan. Traditionally, workflows are executed
	// with a reference launch plan which is why this behavior is the default below.
	if input.TaskID > 0 {
		executionModel.TaskID = input.TaskID
	} else {
		executionModel.LaunchPlanID = input.LaunchPlanID
	}
	if input.RequestSpec.GetMetadata() != nil {
		executionModel.Mode = int32(input.RequestSpec.GetMetadata().GetMode())
	}

	return executionModel, nil
}

// CreateExecutionTagModel transforms a CreateExecutionModelInput to a ExecutionTag model
func CreateExecutionTagModel(input CreateExecutionModelInput) ([]*models.ExecutionTag, error) {
	tags := make([]*models.ExecutionTag, 0)

	if input.RequestSpec.GetLabels() != nil {
		for k, v := range input.RequestSpec.GetLabels().GetValues() {
			tags = append(tags, &models.ExecutionTag{
				ExecutionKey: models.ExecutionKey{
					Project: input.WorkflowExecutionID.GetProject(),
					Domain:  input.WorkflowExecutionID.GetDomain(),
					Name:    input.WorkflowExecutionID.GetName(),
				},
				Key:   k,
				Value: v,
			})
		}
	}

	for _, v := range input.RequestSpec.GetTags() {
		tags = append(tags, &models.ExecutionTag{
			ExecutionKey: models.ExecutionKey{
				Project: input.WorkflowExecutionID.GetProject(),
				Domain:  input.WorkflowExecutionID.GetDomain(),
				Name:    input.WorkflowExecutionID.GetName(),
			},
			Key:   v,
			Value: "",
		})
	}

	return tags, nil
}

func reassignCluster(ctx context.Context, cluster string, executionID *core.WorkflowExecutionIdentifier, execution *models.Execution) error {
	logger.Debugf(ctx, "Updating cluster for execution [%v] with existing recorded cluster [%s] and setting to cluster [%s]",
		executionID, execution.Cluster, cluster)
	execution.Cluster = cluster
	var executionSpec admin.ExecutionSpec
	err := proto.Unmarshal(execution.Spec, &executionSpec)
	if err != nil {
		return flyteErrs.NewFlyteAdminErrorf(codes.Internal, "Failed to unmarshal execution spec: %v", err)
	}
	if executionSpec.GetMetadata() == nil {
		executionSpec.Metadata = &admin.ExecutionMetadata{}
	}
	if executionSpec.GetMetadata().GetSystemMetadata() == nil {
		executionSpec.Metadata.SystemMetadata = &admin.SystemMetadata{}
	}
	executionSpec.Metadata.SystemMetadata.ExecutionCluster = cluster
	marshaledSpec, err := proto.Marshal(&executionSpec)
	if err != nil {
		return flyteErrs.NewFlyteAdminErrorf(codes.Internal, "Failed to marshal execution spec: %v", err)
	}
	execution.Spec = marshaledSpec
	return nil
}

// Updates an existing model given a WorkflowExecution event.
func UpdateExecutionModelState(
	ctx context.Context,
	execution *models.Execution, request *admin.WorkflowExecutionEventRequest,
	inlineEventDataPolicy interfaces.InlineEventDataPolicy, storageClient *storage.DataStore) error {
	var executionClosure admin.ExecutionClosure
	err := proto.Unmarshal(execution.Closure, &executionClosure)
	if err != nil {
		return flyteErrs.NewFlyteAdminErrorf(codes.Internal, "Failed to unmarshal execution closure: %v", err)
	}
	executionClosure.Phase = request.GetEvent().GetPhase()
	executionClosure.UpdatedAt = request.GetEvent().GetOccurredAt()
	execution.Phase = request.GetEvent().GetPhase().String()

	occurredAtTimestamp, err := ptypes.Timestamp(request.GetEvent().GetOccurredAt())
	if err != nil {
		return flyteErrs.NewFlyteAdminErrorf(codes.Internal, "Failed to parse OccurredAt: %v", err)
	}
	execution.ExecutionUpdatedAt = &occurredAtTimestamp

	// only mark the execution started when we get the initial running event
	if request.GetEvent().GetPhase() == core.WorkflowExecution_RUNNING {
		execution.StartedAt = &occurredAtTimestamp
		executionClosure.StartedAt = request.GetEvent().GetOccurredAt()
	} else if common.IsExecutionTerminal(request.GetEvent().GetPhase()) {
		if execution.StartedAt != nil {
			execution.Duration = occurredAtTimestamp.Sub(*execution.StartedAt)
			executionClosure.Duration = ptypes.DurationProto(execution.Duration)
		} else {
			logger.Infof(context.Background(),
				"Cannot compute duration because startedAt was never set, requestId: %v", request.GetRequestId())
		}
	}

	// Default or empty cluster values do not require updating the execution model.
	ignoreClusterFromEvent := len(request.GetEvent().GetProducerId()) == 0 || request.GetEvent().GetProducerId() == common.DefaultProducerID
	logger.Debugf(ctx, "Producer Id [%v]. IgnoreClusterFromEvent [%v]", request.GetEvent().GetProducerId(), ignoreClusterFromEvent)
	if !ignoreClusterFromEvent {
		if clusterReassignablePhases.Has(execution.Phase) {
			if err := reassignCluster(ctx, request.GetEvent().GetProducerId(), request.GetEvent().GetExecutionId(), execution); err != nil {
				return err
			}
		} else if execution.Cluster != request.GetEvent().GetProducerId() {
			errorMsg := fmt.Sprintf("Cannot accept events for running/terminated execution [%v] from cluster [%s],"+
				"expected events to originate from [%s]",
				request.GetEvent().GetExecutionId(), request.GetEvent().GetProducerId(), execution.Cluster)
			return flyteErrs.NewIncompatibleClusterError(ctx, errorMsg, execution.Cluster)
		}
	}

	if request.GetEvent().GetOutputUri() != "" {
		executionClosure.OutputResult = &admin.ExecutionClosure_Outputs{
			Outputs: &admin.LiteralMapBlob{
				Data: &admin.LiteralMapBlob_Uri{
					Uri: request.GetEvent().GetOutputUri(),
				},
			},
		}
	} else if request.GetEvent().GetOutputData() != nil {
		switch inlineEventDataPolicy {
		case interfaces.InlineEventDataPolicyStoreInline:
			executionClosure.OutputResult = &admin.ExecutionClosure_OutputData{
				OutputData: request.GetEvent().GetOutputData(),
			}
		default:
			logger.Debugf(ctx, "Offloading outputs per InlineEventDataPolicy")
			uri, err := common.OffloadLiteralMap(ctx, storageClient, request.GetEvent().GetOutputData(),
				request.GetEvent().GetExecutionId().GetProject(), request.GetEvent().GetExecutionId().GetDomain(), request.GetEvent().GetExecutionId().GetName(), OutputsObjectSuffix)
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
	} else if request.GetEvent().GetError() != nil {
		executionClosure.OutputResult = &admin.ExecutionClosure_Error{
			Error: request.GetEvent().GetError(),
		}
		k := request.GetEvent().GetError().GetKind().String()
		execution.ErrorKind = &k
		execution.ErrorCode = &request.Event.GetError().Code
	}
	marshaledClosure, err := proto.Marshal(&executionClosure)
	if err != nil {
		return flyteErrs.NewFlyteAdminErrorf(codes.Internal, "Failed to marshal execution closure: %v", err)
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
		return flyteErrs.NewFlyteAdminErrorf(codes.Internal, "Failed to unmarshal execution closure: %v", err)
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
		return flyteErrs.NewFlyteAdminErrorf(codes.Internal, "Failed to marshal execution closure: %v", err)
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
		return flyteErrs.NewFlyteAdminErrorf(codes.Internal, "Failed to unmarshal execution closure: %v", err)
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
		return flyteErrs.NewFlyteAdminErrorf(codes.Internal, "Failed to marshal execution closure: %v", err)
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
		return nil, flyteErrs.NewFlyteAdminErrorf(codes.Internal, "failed to unmarshal spec")
	}
	if len(opts.DefaultNamespace) > 0 {
		if spec.GetMetadata() == nil {
			spec.Metadata = &admin.ExecutionMetadata{}
		}
		if spec.GetMetadata().GetSystemMetadata() == nil {
			spec.Metadata.SystemMetadata = &admin.SystemMetadata{}
		}
		if len(spec.GetMetadata().GetSystemMetadata().GetNamespace()) == 0 {
			logger.Infof(ctx, "setting execution system metadata namespace to [%s]", opts.DefaultNamespace)
			spec.Metadata.SystemMetadata.Namespace = opts.DefaultNamespace
		}
	}

	var closure admin.ExecutionClosure
	if err = proto.Unmarshal(executionModel.Closure, &closure); err != nil {
		return nil, flyteErrs.NewFlyteAdminErrorf(codes.Internal, "failed to unmarshal closure")
	}
	if closure.GetError() != nil && opts != nil && opts.TrimErrorMessage && len(closure.GetError().GetMessage()) > 0 {
		trimmedErrOutputResult := closure.GetError()
		trimmedErrMessage := TrimErrorMessage(trimmedErrOutputResult.GetMessage())
		trimmedErrOutputResult.Message = trimmedErrMessage
		closure.OutputResult = &admin.ExecutionClosure_Error{
			Error: trimmedErrOutputResult,
		}
	}

	if closure.GetStateChangeDetails() == nil {
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

// PopulateDefaultStateChangeDetails used to populate execution state change details for older executions which do not
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
