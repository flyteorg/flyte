package k8s

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
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
		k8sClient: fake.NewClientBuilder().WithScheme(scheme).WithObjects(parent).Build(),
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

	assert.Nil(t, buildActionUpdate(ctx, ta, watch.Modified).Value)

	lit := &core.Literal{
		Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Primitive{Primitive: &core.Primitive{
				Value: &core.Primitive_Boolean{Boolean: true},
			}},
		}},
	}
	raw, err := proto.Marshal(lit)
	require.NoError(t, err)
	ta.Status.SignalValue = raw

	value := buildActionUpdate(ctx, ta, watch.Modified).Value
	require.NotNil(t, value)
	assert.True(t, value.GetScalar().GetPrimitive().GetBoolean())
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
