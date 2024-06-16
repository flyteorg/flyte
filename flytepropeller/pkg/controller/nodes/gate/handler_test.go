package gate

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/durationpb"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	ioMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	flyteMocks "github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	executormocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/gate/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/handler"
	nodeMocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces/mocks"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

var (
	eventConfig = &config.EventConfig{
		RawOutputPolicy: config.RawOutputPolicyReference,
	}

	approveGateNode = &v1alpha1.GateNodeSpec{
		Kind: v1alpha1.ConditionKindApprove,
		Approve: &v1alpha1.ApproveCondition{
			ApproveCondition: &core.ApproveCondition{
				SignalId: "foo",
			},
		},
	}

	signalGateNode = &v1alpha1.GateNodeSpec{
		Kind: v1alpha1.ConditionKindSignal,
		Signal: &v1alpha1.SignalCondition{
			SignalCondition: &core.SignalCondition{
				SignalId: "foo",
				Type: &core.LiteralType{
					Type: &core.LiteralType_Simple{
						Simple: core.SimpleType_BOOLEAN,
					},
				},
				OutputVariableName: "bar",
			},
		},
	}

	sleepMinuteGateNode = &v1alpha1.GateNodeSpec{
		Kind: v1alpha1.ConditionKindSleep,
		Sleep: &v1alpha1.SleepCondition{
			SleepCondition: &core.SleepCondition{
				Duration: durationpb.New(time.Minute),
			},
		},
	}

	sleepNowGateNode = &v1alpha1.GateNodeSpec{
		Kind: v1alpha1.ConditionKindSleep,
		Sleep: &v1alpha1.SleepCondition{
			SleepCondition: &core.SleepCondition{
				Duration: durationpb.New(time.Minute * 0),
			},
		},
	}
)

func init() {
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey,
		contextutils.TaskIDKey)
}

func createNodeExecutionContext(gateNode *v1alpha1.GateNodeSpec) *nodeMocks.NodeExecutionContext {
	wfExecID := v1alpha1.WorkflowExecutionIdentifier{
		WorkflowExecutionIdentifier: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
	}

	n := &flyteMocks.ExecutableNode{}
	n.OnGetGateNode().Return(gateNode)

	nm := &nodeMocks.NodeExecutionMetadata{}

	ns := &flyteMocks.ExecutableNodeStatus{}
	ns.OnGetDataDir().Return(storage.DataReference("data-dir"))
	ns.OnGetOutputDir().Return(storage.DataReference("data-dir"))

	t := v1.NewTime(time.Now())
	ns.OnGetLastAttemptStartedAt().Return(&t)

	inputReader := &ioMocks.InputReader{}
	inputReader.OnGetMatch(mock.Anything).Return(&core.LiteralMap{}, nil)
	dataStore, _ := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())

	eCtx := &executormocks.ExecutionContext{}
	eCtx.EXPECT().GetExecutionID().Return(wfExecID)

	nCtx := &nodeMocks.NodeExecutionContext{}
	nCtx.OnNodeExecutionMetadata().Return(nm)
	nCtx.OnNode().Return(n)
	nCtx.OnNodeStatus().Return(ns)
	nCtx.OnDataStore().Return(dataStore)
	nCtx.OnExecutionContext().Return(eCtx)
	nCtx.OnInputReader().Return(inputReader)

	r := &nodeMocks.NodeStateReader{}
	r.OnGetGateNodeState().Return(handler.GateNodeState{})
	nCtx.OnNodeStateReader().Return(r)

	w := &nodeMocks.NodeStateWriter{}
	w.OnPutGateNodeStateMatch(mock.Anything).Return(nil)
	nCtx.OnNodeStateWriter().Return(w)
	return nCtx
}

func TestAbort(t *testing.T) {
	ctx := context.TODO()
	signalClient := mocks.SignalServiceClient{}
	scope := promutils.NewTestScope()

	handler := New(eventConfig, &signalClient, scope)

	assert.NoError(t, handler.Abort(ctx, nil, ""))
}

func TestFinalize(t *testing.T) {
	ctx := context.TODO()
	signalClient := mocks.SignalServiceClient{}
	scope := promutils.NewTestScope()

	handler := New(eventConfig, &signalClient, scope)

	assert.NoError(t, handler.Finalize(ctx, nil))
}

func TestHandle(t *testing.T) {
	ctx := context.TODO()
	scope := promutils.NewTestScope()

	t.Run("ApproveCheck", func(t *testing.T) {
		nCtx := createNodeExecutionContext(approveGateNode)
		signalClient := mocks.SignalServiceClient{}
		signalClient.OnGetOrCreateSignalMatch(mock.Anything, mock.Anything).Return(&admin.Signal{}, nil)

		gateNodeHandler := New(eventConfig, &signalClient, scope)

		transition, err := gateNodeHandler.Handle(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, handler.EPhaseRunning, transition.Info().GetPhase())
	})

	t.Run("ApproveComplete", func(t *testing.T) {
		nCtx := createNodeExecutionContext(approveGateNode)
		signalClient := mocks.SignalServiceClient{}
		signalClient.OnGetOrCreateSignalMatch(mock.Anything, mock.Anything).Return(&admin.Signal{
			Value: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Primitive{
							Primitive: &core.Primitive{
								Value: &core.Primitive_Boolean{
									Boolean: true,
								},
							},
						},
					},
				},
			},
		}, nil)

		gateNodeHandler := New(eventConfig, &signalClient, scope)

		transition, err := gateNodeHandler.Handle(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, handler.EPhaseSuccess, transition.Info().GetPhase())
	})

	t.Run("ApproveRejected", func(t *testing.T) {
		nCtx := createNodeExecutionContext(approveGateNode)
		signalClient := mocks.SignalServiceClient{}
		signalClient.OnGetOrCreateSignalMatch(mock.Anything, mock.Anything).Return(&admin.Signal{
			Value: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Primitive{
							Primitive: &core.Primitive{
								Value: &core.Primitive_Boolean{
									Boolean: false,
								},
							},
						},
					},
				},
			},
		}, nil)

		gateNodeHandler := New(eventConfig, &signalClient, scope)

		transition, err := gateNodeHandler.Handle(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, handler.EPhaseFailed, transition.Info().GetPhase())
	})

	t.Run("ApproveError", func(t *testing.T) {
		nCtx := createNodeExecutionContext(approveGateNode)
		signalClient := mocks.SignalServiceClient{}
		signalClient.OnGetOrCreateSignalMatch(mock.Anything, mock.Anything).Return(&admin.Signal{}, errors.New("foo"))

		gateNodeHandler := New(eventConfig, &signalClient, scope)

		transition, err := gateNodeHandler.Handle(ctx, nCtx)
		assert.Error(t, err)
		assert.Equal(t, handler.EPhaseUndefined, transition.Info().GetPhase())
	})

	t.Run("SignalCheck", func(t *testing.T) {
		nCtx := createNodeExecutionContext(signalGateNode)
		signalClient := mocks.SignalServiceClient{}
		signalClient.OnGetOrCreateSignalMatch(mock.Anything, mock.Anything).Return(&admin.Signal{}, nil)

		gateNodeHandler := New(eventConfig, &signalClient, scope)

		transition, err := gateNodeHandler.Handle(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, handler.EPhaseRunning, transition.Info().GetPhase())
	})

	t.Run("SignalComplete", func(t *testing.T) {
		nCtx := createNodeExecutionContext(signalGateNode)
		signalClient := mocks.SignalServiceClient{}
		signalClient.OnGetOrCreateSignalMatch(mock.Anything, mock.Anything).Return(&admin.Signal{
			Value: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Primitive{
							Primitive: &core.Primitive{
								Value: &core.Primitive_Boolean{
									Boolean: false,
								},
							},
						},
					},
				},
			},
		}, nil)

		gateNodeHandler := New(eventConfig, &signalClient, scope)

		transition, err := gateNodeHandler.Handle(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, handler.EPhaseSuccess, transition.Info().GetPhase())
	})

	t.Run("SignalError", func(t *testing.T) {
		nCtx := createNodeExecutionContext(signalGateNode)
		signalClient := mocks.SignalServiceClient{}
		signalClient.OnGetOrCreateSignalMatch(mock.Anything, mock.Anything).Return(&admin.Signal{}, errors.New("foo"))

		gateNodeHandler := New(eventConfig, &signalClient, scope)

		transition, err := gateNodeHandler.Handle(ctx, nCtx)
		assert.Error(t, err)
		assert.Equal(t, handler.EPhaseUndefined, transition.Info().GetPhase())
	})

	t.Run("SleepCheck", func(t *testing.T) {
		nCtx := createNodeExecutionContext(sleepMinuteGateNode)
		signalClient := mocks.SignalServiceClient{}

		gateNodeHandler := New(eventConfig, &signalClient, scope)

		transition, err := gateNodeHandler.Handle(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, handler.EPhaseRunning, transition.Info().GetPhase())
	})

	t.Run("SleepComplete", func(t *testing.T) {
		nCtx := createNodeExecutionContext(sleepNowGateNode)
		signalClient := mocks.SignalServiceClient{}

		gateNodeHandler := New(eventConfig, &signalClient, scope)

		transition, err := gateNodeHandler.Handle(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, handler.EPhaseSuccess, transition.Info().GetPhase())
	})
}
