package gate

import (
	"context"
	"fmt"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/errors"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/interfaces"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/storage"
)

//go:generate mockery -all -case=underscore

// SignalServiceClient is a SignalServiceClient wrapper interface used specifically for generating
// mocks for testing
type SignalServiceClient interface {
	service.SignalServiceClient
}

// gateNodeHandler is a handle implementation for processing gate nodes
type gateNodeHandler struct {
	signalClient SignalServiceClient
	metrics      metrics
}

// metrics encapsulates the prometheus metrics for this handler
type metrics struct {
	scope promutils.Scope
}

// newMetrics initializes a new metrics struct
func newMetrics(scope promutils.Scope) metrics {
	return metrics{
		scope: scope,
	}
}

// Abort stops the gate node defined in the NodeExecutionContext
func (g *gateNodeHandler) Abort(ctx context.Context, nCtx interfaces.NodeExecutionContext, reason string) error {
	return nil
}

// Finalize completes the gate node defined in the NodeExecutionContext
func (g *gateNodeHandler) Finalize(ctx context.Context, _ interfaces.NodeExecutionContext) error {
	return nil
}

// FinalizeRequired defines whether or not this handler requires finalize to be called on
// node completion
func (g *gateNodeHandler) FinalizeRequired() bool {
	return false
}

// Handle is responsible for transitioning and reporting node state to complete the node defined
// by the NodeExecutionContext
func (g *gateNodeHandler) Handle(ctx context.Context, nCtx interfaces.NodeExecutionContext) (handler.Transition, error) {
	gateNode := nCtx.Node().GetGateNode()
	gateNodeState := nCtx.NodeStateReader().GetGateNodeState()

	if gateNodeState.Phase == v1alpha1.GateNodePhaseUndefined {
		gateNodeState.Phase = v1alpha1.GateNodePhaseExecuting
	}

	switch gateNode.GetKind() {
	case v1alpha1.ConditionKindApprove:
		// retrieve approve condition
		approveCondition := gateNode.GetApprove()
		if approveCondition == nil {
			errMsg := "gateNode approve condition is nil"
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM,
				errors.BadSpecificationError, errMsg, nil)), nil
		}

		// use admin client to query for signal
		request := &admin.SignalGetOrCreateRequest{
			Id: &core.SignalIdentifier{
				ExecutionId: nCtx.ExecutionContext().GetExecutionID().WorkflowExecutionIdentifier,
				SignalId:    approveCondition.SignalId,
			},
			Type: &core.LiteralType{
				Type: &core.LiteralType_Simple{
					Simple: core.SimpleType_BOOLEAN,
				},
			},
		}

		signal, err := g.signalClient.GetOrCreateSignal(ctx, request)
		if err != nil {
			return handler.UnknownTransition, err
		}

		// if signal has value then check for approval
		if signal.Value != nil && signal.Value.Value != nil {
			approved, ok := getBoolean(signal.Value)
			if !ok {
				errMsg := fmt.Sprintf("received a non-boolean approve signal value [%v]", signal.Value)
				return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_UNKNOWN,
					errors.RuntimeExecutionError, errMsg, nil)), nil
			}

			if !approved {
				return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_USER,
					"ReceivedRejectSignal", "received a reject signal to disapprove the node input values", nil)), nil
			}

			// copy input values to outputs
			inputs, err := nCtx.InputReader().Get(ctx)
			if err != nil {
				errMsg := fmt.Sprintf("failed to read input with error [%s]", err)
				return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM, errors.RuntimeExecutionError, errMsg, nil)), nil
			}

			outputFile := v1alpha1.GetOutputsFile(nCtx.NodeStatus().GetOutputDir())
			if err := nCtx.DataStore().WriteProtobuf(ctx, outputFile, storage.Options{}, inputs); err != nil {
				return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM, "WriteOutputsFailed",
					fmt.Sprintf("failed to write signal value to [%v] with error [%s]", outputFile, err.Error()), nil)), nil
			}

			o := &handler.OutputInfo{OutputURI: outputFile}
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoSuccess(&handler.ExecutionInfo{
				OutputInfo: o,
			})), nil
		}
	case v1alpha1.ConditionKindSignal:
		// retrieve signal condition
		signalCondition := gateNode.GetSignal()
		if signalCondition == nil {
			errMsg := "gateNode signal condition is nil"
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM,
				errors.BadSpecificationError, errMsg, nil)), nil
		}

		// use admin client to query for signal
		request := &admin.SignalGetOrCreateRequest{
			Id: &core.SignalIdentifier{
				ExecutionId: nCtx.ExecutionContext().GetExecutionID().WorkflowExecutionIdentifier,
				SignalId:    signalCondition.SignalId,
			},
			Type: signalCondition.Type,
		}

		signal, err := g.signalClient.GetOrCreateSignal(ctx, request)
		if err != nil {
			return handler.UnknownTransition, err
		}

		// if signal has value then write to output and transition to success
		if signal.Value != nil && signal.Value.Value != nil {
			outputs := &core.LiteralMap{
				Literals: map[string]*core.Literal{
					signalCondition.OutputVariableName: signal.Value,
				},
			}

			outputFile := v1alpha1.GetOutputsFile(nCtx.NodeStatus().GetOutputDir())
			if err := nCtx.DataStore().WriteProtobuf(ctx, outputFile, storage.Options{}, outputs); err != nil {
				return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM, "WriteOutputsFailed",
					fmt.Sprintf("failed to write signal value to [%v] with error: %s", outputFile, err.Error()), nil)), nil
			}

			o := &handler.OutputInfo{OutputURI: outputFile}
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoSuccess(&handler.ExecutionInfo{
				OutputInfo: o,
			})), nil
		}
	case v1alpha1.ConditionKindSleep:
		// retrieve sleep duration
		sleepCondition := gateNode.GetSleep()
		if sleepCondition == nil {
			errMsg := "gateNode sleep condition is nil"
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM,
				errors.BadSpecificationError, errMsg, nil)), nil
		}

		sleepDuration := sleepCondition.GetDuration().AsDuration()

		// check duration of node sleep
		lastAttemptStartedAt := nCtx.NodeStatus().GetLastAttemptStartedAt()
		if lastAttemptStartedAt != nil && sleepDuration <= time.Since(lastAttemptStartedAt.Time) {
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoSuccess(&handler.ExecutionInfo{})), nil
		}
	default:
		errMsg := "gateNode does not have a supported condition reference"
		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_SYSTEM,
			errors.BadSpecificationError, errMsg, nil)), nil
	}

	// update gate node status
	if err := nCtx.NodeStateWriter().PutGateNodeState(gateNodeState); err != nil {
		logger.Errorf(ctx, "failed to store GateNode state with err [%s]", err.Error())
		return handler.UnknownTransition, err
	}

	return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRunning(&handler.ExecutionInfo{})), nil
}

// Setup handles any initialization requirements for this handler
func (g *gateNodeHandler) Setup(_ context.Context, _ interfaces.SetupContext) error {
	return nil
}

// New initializes a new gateNodeHandler
func New(eventConfig *config.EventConfig, signalClient service.SignalServiceClient, scope promutils.Scope) interfaces.NodeHandler {
	gateScope := scope.NewSubScope("gate")
	return &gateNodeHandler{
		signalClient: signalClient,
		metrics:      newMetrics(gateScope),
	}
}

func getBoolean(literal *core.Literal) (bool, bool) {
	if scalarValue, ok := literal.Value.(*core.Literal_Scalar); ok {
		if primitiveValue, ok := scalarValue.Scalar.Value.(*core.Scalar_Primitive); ok {
			if booleanValue, ok := primitiveValue.Primitive.Value.(*core.Primitive_Boolean); ok {
				return booleanValue.Boolean, true
			}
		}
	}

	return false, false
}
