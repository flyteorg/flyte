package k8s

import (
	"context"
	"testing"

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
// IDs; the second Enqueue must succeed and leave the existing CR untouched.
func TestEnqueueTask_AlreadyExistsIsIdempotent(t *testing.T) {
	ctx := context.Background()
	c := newTaskClient(t, nil)
	action := newChildTaskAction()

	require.NoError(t, c.Enqueue(ctx, action, nil))
	created, err := c.GetTaskAction(ctx, action.ActionId)
	require.NoError(t, err)

	assert.NoError(t, c.Enqueue(ctx, action, nil))

	existing, err := c.GetTaskAction(ctx, action.ActionId)
	require.NoError(t, err)
	assert.Equal(t, created.ResourceVersion, existing.ResourceVersion)
}

func TestEnqueueTask_CreateError(t *testing.T) {
	gr := schema.GroupResource{Group: "flyte.org", Resource: "taskactions"}

	for _, tc := range []struct {
		name      string
		createErr error
		wantErr   bool
	}{
		{name: "already exists is ignored", createErr: apierrors.NewAlreadyExists(gr, "run1-a1")},
		{name: "other errors are returned", createErr: apierrors.NewForbidden(gr, "run1-a1", nil), wantErr: true},
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
			if !tc.wantErr {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.True(t, apierrors.IsForbidden(err), "expected wrapped Forbidden error, got %v", err)
		})
	}
}
