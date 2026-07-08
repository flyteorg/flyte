package k8s

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	executorv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	runmocks "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect/mocks"
)

// newConditionClient returns an ActionsClient backed by a fake k8s client
// pre-loaded with a parent TaskAction "run1-a0".
func newConditionClient(t *testing.T) *ActionsClient {
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme)) // EnsureNamespaceExists needs v1.Namespace
	require.NoError(t, executorv1.AddToScheme(scheme))
	parent := &executorv1.TaskAction{
		ObjectMeta: metav1.ObjectMeta{Name: "run1-a0", Namespace: flyteNamespace},
	}
	return &ActionsClient{
		k8sClient: fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(parent).
			WithStatusSubresource(&executorv1.TaskAction{}).
			Build(),
	}
}

func boolLiteral(v bool) *core.Literal {
	return &core.Literal{
		Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Primitive{Primitive: &core.Primitive{
				Value: &core.Primitive_Boolean{Boolean: v},
			}},
		}},
	}
}

func strLiteral(s string) *core.Literal {
	return &core.Literal{
		Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Primitive{Primitive: &core.Primitive{
				Value: &core.Primitive_StringValue{StringValue: s},
			}},
		}},
	}
}

func newConditionAction(simpleType core.SimpleType) *actions.Action {
	parent := "a0"
	return &actions.Action{
		ActionId: &common.ActionIdentifier{
			Run:  &common.RunIdentifier{Project: "proj", Domain: "dev", Name: "run1"},
			Name: "cond1",
		},
		ParentActionName: &parent,
		Spec: &actions.Action_Condition{
			Condition: &workflow.ConditionAction{
				Name:    "approve",
				Type:    &core.LiteralType{Type: &core.LiteralType_Simple{Simple: simpleType}},
				Timeout: durationpb.New(time.Hour),
			},
		},
	}
}

func TestEnqueueCondition_CreatesCR(t *testing.T) {
	ctx := context.Background()
	c := newConditionClient(t)
	action := newConditionAction(core.SimpleType_BOOLEAN)

	require.NoError(t, c.Enqueue(ctx, action, nil))

	created, err := c.GetTaskAction(ctx, action.ActionId)
	require.NoError(t, err)
	assert.Equal(t, executorv1.ActionTypeCondition, created.Spec.ActionType)
	assert.Equal(t, "condition", created.Labels["flyte.org/action-type"])
	require.Len(t, created.OwnerReferences, 1)
	assert.Equal(t, "run1-a0", created.OwnerReferences[0].Name)

	condSpec := &workflow.ConditionAction{}
	require.NoError(t, proto.Unmarshal(created.Spec.ConditionSpec, condSpec))
	assert.True(t, proto.Equal(action.GetCondition(), condSpec))
}

func TestEnqueueCondition_InvalidDeclaredType(t *testing.T) {
	err := newConditionClient(t).Enqueue(context.Background(), newConditionAction(core.SimpleType_STRUCT), nil)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

func TestEnqueueCondition_AlreadyExistsIsIdempotent(t *testing.T) {
	ctx := context.Background()
	c := newConditionClient(t)
	action := newConditionAction(core.SimpleType_BOOLEAN)

	require.NoError(t, c.Enqueue(ctx, action, nil))
	assert.NoError(t, c.Enqueue(ctx, action, nil))
}

func TestBuildActionUpdate_SignalValue(t *testing.T) {
	ctx := context.Background()
	ta, _ := newTestActionUpdate("cond1")

	assert.Nil(t, buildActionUpdate(ctx, ta, watch.Modified).SignalValue)

	raw, err := proto.Marshal(boolLiteral(true))
	require.NoError(t, err)
	ta.Status.SignalValue = raw

	value := buildActionUpdate(ctx, ta, watch.Modified).SignalValue
	require.NotNil(t, value)
	assert.True(t, value.GetScalar().GetPrimitive().GetBoolean())
}

func TestSignal(t *testing.T) {
	ctx := context.Background()

	// newSignalledClient returns a client with the condition "cond1" enqueued.
	newSignalledClient := func(t *testing.T) *ActionsClient {
		c := newConditionClient(t)
		require.NoError(t, c.Enqueue(ctx, newConditionAction(core.SimpleType_BOOLEAN), nil))
		return c
	}
	condID := newConditionAction(core.SimpleType_BOOLEAN).ActionId

	t.Run("happy path persists signal fields", func(t *testing.T) {
		c := newSignalledClient(t)
		require.NoError(t, c.Signal(ctx, condID, boolLiteral(true), "user@example.com"))

		ta, err := c.GetTaskAction(ctx, condID)
		require.NoError(t, err)
		assert.True(t, proto.Equal(boolLiteral(true), SignalValueFromStatus(ctx, ta)))
		assert.Equal(t, "user@example.com", ta.Status.SignalledBy)
		assert.NotNil(t, ta.Status.SignalledAt)
	})

	t.Run("empty literal is invalid", func(t *testing.T) {
		err := newSignalledClient(t).Signal(ctx, condID, &core.Literal{}, "u")
		assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
	})

	t.Run("not found", func(t *testing.T) {
		missing := &common.ActionIdentifier{Run: condID.Run, Name: "nope"}
		err := newSignalledClient(t).Signal(ctx, missing, boolLiteral(true), "u")
		assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
	})

	t.Run("not a condition", func(t *testing.T) {
		parentID := &common.ActionIdentifier{Run: condID.Run, Name: "run1"} // resolves to run1-a0, a task CR
		err := newSignalledClient(t).Signal(ctx, parentID, boolLiteral(true), "u")
		assert.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))
	})

	t.Run("type mismatch", func(t *testing.T) {
		err := newSignalledClient(t).Signal(ctx, condID, strLiteral("yes"), "u")
		assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
	})

	t.Run("re-signal with same value is idempotent", func(t *testing.T) {
		c := newSignalledClient(t)
		require.NoError(t, c.Signal(ctx, condID, boolLiteral(true), "u"))
		assert.NoError(t, c.Signal(ctx, condID, boolLiteral(true), "u"))
	})

	t.Run("re-signal with different value fails", func(t *testing.T) {
		c := newSignalledClient(t)
		require.NoError(t, c.Signal(ctx, condID, boolLiteral(true), "u"))
		err := c.Signal(ctx, condID, boolLiteral(false), "u")
		assert.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))
	})

	t.Run("already terminal", func(t *testing.T) {
		c := newSignalledClient(t)
		ta, err := c.GetTaskAction(ctx, condID)
		require.NoError(t, err)
		ta.Status.Conditions = []metav1.Condition{{
			Type:               string(executorv1.ConditionTypeFailed),
			Status:             metav1.ConditionTrue,
			Reason:             string(executorv1.ConditionReasonTimedOut),
			LastTransitionTime: metav1.Now(),
		}}
		require.NoError(t, c.k8sClient.Status().Update(ctx, ta))

		err = c.Signal(ctx, condID, boolLiteral(true), "u")
		assert.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))
	})
}

func TestNotifyRunService_Condition(t *testing.T) {
	ctx := context.Background()

	newClientWithRunMock := func(t *testing.T) (*runmocks.InternalRunServiceClient, *ActionsClient) {
		mockClient := runmocks.NewInternalRunServiceClient(t)
		return mockClient, &ActionsClient{
			runClient:   mockClient,
			subscribers: make(map[string]map[chan *ActionUpdate]struct{}),
		}
	}

	condSpecBytes, err := proto.Marshal(&workflow.ConditionAction{Name: "approve"})
	require.NoError(t, err)

	t.Run("added event records condition spec", func(t *testing.T) {
		mockClient, c := newClientWithRunMock(t)
		ta, update := newTestActionUpdate("cond1")
		ta.Spec.ActionType = executorv1.ActionTypeCondition
		ta.Spec.ConditionSpec = condSpecBytes

		mockClient.EXPECT().RecordAction(mock.Anything, mock.MatchedBy(func(r *connect.Request[workflow.RecordActionRequest]) bool {
			return r.Msg.GetCondition().GetName() == "approve"
		})).Return(&connect.Response[workflow.RecordActionResponse]{}, nil)

		c.notifyRunService(ctx, ta, update, watch.Added)
	})

	t.Run("terminal succeeded carries output and principal", func(t *testing.T) {
		mockClient, c := newClientWithRunMock(t)
		ta, update := newTestActionUpdate("cond1")
		ta.Spec.ActionType = executorv1.ActionTypeCondition
		ta.Status.SignalledBy = "user@example.com"
		update.Phase = common.ActionPhase_ACTION_PHASE_SUCCEEDED
		update.SignalValue = boolLiteral(true)

		mockClient.EXPECT().UpdateActionStatus(mock.Anything, mock.MatchedBy(func(r *connect.Request[workflow.UpdateActionStatusRequest]) bool {
			return proto.Equal(r.Msg.GetOutput(), boolLiteral(true)) &&
				r.Msg.GetPrincipal().GetUser().GetId().GetSubject() == "user@example.com"
		})).Return(&connect.Response[workflow.UpdateActionStatusResponse]{}, nil)

		c.notifyRunService(ctx, ta, update, watch.Modified)
	})
}

func TestGetPhaseFromConditions_ConditionPhases(t *testing.T) {
	tests := []struct {
		name      string
		condition metav1.Condition
		want      common.ActionPhase
	}{
		{
			name: "paused",
			condition: metav1.Condition{
				Type:   string(executorv1.ConditionTypeProgressing),
				Status: metav1.ConditionTrue,
				Reason: string(executorv1.ConditionReasonPaused),
			},
			want: common.ActionPhase_ACTION_PHASE_PAUSED,
		},
		{
			name: "timed out",
			condition: metav1.Condition{
				Type:   string(executorv1.ConditionTypeFailed),
				Status: metav1.ConditionTrue,
				Reason: string(executorv1.ConditionReasonTimedOut),
			},
			want: common.ActionPhase_ACTION_PHASE_TIMED_OUT,
		},
		{
			name: "plain failure keeps FAILED",
			condition: metav1.Condition{
				Type:   string(executorv1.ConditionTypeFailed),
				Status: metav1.ConditionTrue,
				Reason: string(executorv1.ConditionReasonPermanentFailure),
			},
			want: common.ActionPhase_ACTION_PHASE_FAILED,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := &executorv1.TaskAction{
				Status: executorv1.TaskActionStatus{Conditions: []metav1.Condition{tt.condition}},
			}
			assert.Equal(t, tt.want, GetPhaseFromConditions(ta))
		})
	}
}
