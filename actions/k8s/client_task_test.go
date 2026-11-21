package k8s

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	executorv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

// newTaskClient returns an ActionsClient backed by a fake k8s client
// pre-loaded with a parent TaskAction "run1-a0". funcs, if non-nil, intercepts
// k8s calls so tests can inject errors.
func newTaskClient(t *testing.T, funcs *interceptor.Funcs) *ActionsClient {
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme)) // EnsureNamespaceExists needs v1.Namespace
	require.NoError(t, executorv1.AddToScheme(scheme))
	parent := &executorv1.TaskAction{
		ObjectMeta: metav1.ObjectMeta{Name: "run1-a0", Namespace: "flyte"},
	}
	builder := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(parent).
		WithStatusSubresource(&executorv1.TaskAction{})
	if funcs != nil {
		builder = builder.WithInterceptorFuncs(*funcs)
	}
	return &ActionsClient{
		recordedFilter: testFilter(),
		namespace:      "flyte",
		k8sClient:      builder.Build(),
	}
}

// newChildTaskAction returns a task action "a1" whose parent is "a0".
func newChildTaskAction() *actions.Action {
	parent := "a0"
	return &actions.Action{
		ActionId: &common.ActionIdentifier{
			Run:  &common.RunIdentifier{Project: "proj", Domain: "dev", Name: "run1"},
			Name: "a1",
		},
		ParentActionName: &parent,
		Spec: &actions.Action_Task{
			Task: &workflow.TaskAction{
				Spec: &task.TaskSpec{
					TaskTemplate: &core.TaskTemplate{Type: "python"},
				},
			},
		},
	}
}

// A restarted parent re-enqueues children with the same deterministic action
// IDs; the second Enqueue must return AlreadyExists and leave the existing CR
// untouched.
func TestEnqueueTask_AlreadyExists(t *testing.T) {
	ctx := context.Background()
	c := newTaskClient(t, nil)
	action := newChildTaskAction()

	require.NoError(t, c.Enqueue(ctx, action, nil))
	created, err := c.GetTaskAction(ctx, action.ActionId)
	require.NoError(t, err)

	err = c.Enqueue(ctx, action, nil)
	assert.Equal(t, connect.CodeAlreadyExists, connect.CodeOf(err))

	existing, err := c.GetTaskAction(ctx, action.ActionId)
	require.NoError(t, err)
	assert.Equal(t, created.ResourceVersion, existing.ResourceVersion)
}

func TestEnqueueTask_CreateError(t *testing.T) {
	gr := schema.GroupResource{Group: "flyte.org", Resource: "taskactions"}

	for _, tc := range []struct {
		name      string
		createErr error
		check     func(t *testing.T, err error)
	}{
		{
			name:      "already exists maps to CodeAlreadyExists",
			createErr: apierrors.NewAlreadyExists(gr, "run1-a1"),
			check: func(t *testing.T, err error) {
				assert.Equal(t, connect.CodeAlreadyExists, connect.CodeOf(err))
				assert.True(t, apierrors.IsAlreadyExists(err), "expected wrapped AlreadyExists error, got %v", err)
			},
		},
		{
			name:      "other errors are returned unmapped",
			createErr: apierrors.NewForbidden(gr, "run1-a1", nil),
			check: func(t *testing.T, err error) {
				var connectErr *connect.Error
				assert.False(t, errors.As(err, &connectErr), "expected non-connect error, got %v", err)
				assert.True(t, apierrors.IsForbidden(err), "expected wrapped Forbidden error, got %v", err)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newTaskClient(t, &interceptor.Funcs{
				Create: func(ctx context.Context, cl client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
					if _, ok := obj.(*executorv1.TaskAction); ok {
						return tc.createErr
					}
					return cl.Create(ctx, obj, opts...)
				},
			})

			err := c.Enqueue(context.Background(), newChildTaskAction(), nil)
			require.Error(t, err)
			tc.check(t, err)
		})
	}
}

// Children are enqueued without a RunSpec, so the start time must survive in
// the parent CR and propagate through every generation.
func TestEnqueueTask_InheritsRunStartTime(t *testing.T) {
	for _, withStartTime := range []bool{true, false} {
		name := "legacy run without start time"
		if withStartTime {
			name = "run with start time"
		}
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			c := newTaskClient(t, nil)
			runSpec := &task.RunSpec{}
			expected := runStartTimeTemplateVar
			if withStartTime {
				runSpec.RunStartTime = timestamppb.New(time.Date(2026, 6, 3, 12, 0, 0, 0, time.UTC))
				expected = "2026-06-03T12:00:00Z"
			}
			parent := ""
			for _, actionName := range []string{"root", "child", "grandchild"} {
				action := newChildTaskAction()
				action.ActionId.Name = actionName
				action.ParentActionName = &parent
				tmpl := action.GetTask().GetSpec().GetTaskTemplate()
				tmpl.Target = &core.TaskTemplate_Container{Container: &core.Container{
					Args: []string{"--run-start-time", runStartTimeTemplateVar},
				}}
				var spec *task.RunSpec
				if parent == "" {
					spec = runSpec
				}
				require.NoError(t, c.Enqueue(ctx, action, spec))
				created, err := c.GetTaskAction(ctx, action.ActionId)
				require.NoError(t, err)
				embedded := &core.TaskTemplate{}
				require.NoError(t, proto.Unmarshal(created.Spec.TaskTemplate, embedded))
				assert.Equal(t, []string{"--run-start-time", expected}, embedded.GetContainer().GetArgs())
				assert.Equal(t, runStartTimeTemplateVar, tmpl.GetContainer().GetArgs()[1])
				if withStartTime {
					assert.Equal(t, expected, created.Spec.RunStartTime)
				} else {
					assert.Empty(t, created.Spec.RunStartTime)
				}
				parent = actionName
			}
		})
	}
}
