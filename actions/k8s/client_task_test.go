package k8s

import (
	"context"
	"errors"
	"testing"

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

func TestEnqueue_ParentTaskActionLabel(t *testing.T) {
	for _, actionType := range []string{"task", "condition"} {
		t.Run(actionType, func(t *testing.T) {
			ctx := context.Background()
			c := newTaskClient(t, nil)
			newAction := func() *actions.Action {
				if actionType == "condition" {
					return newConditionAction(core.SimpleType_BOOLEAN)
				}
				return newChildTaskAction()
			}

			root := newAction()
			root.ActionId.Name = "root"
			root.ParentActionName = nil
			require.NoError(t, c.Enqueue(ctx, root, nil))
			rootCR, err := c.GetTaskAction(ctx, root.ActionId)
			require.NoError(t, err)
			assert.NotContains(t, rootCR.Labels, "flyte.org/parent-task-action")
			assert.Equal(t, "true", rootCR.Labels[labelParentLabelEnabled])

			child := newAction()
			child.ParentActionName = &root.ActionId.Name
			require.NoError(t, c.Enqueue(ctx, child, nil))
			childCR, err := c.GetTaskAction(ctx, child.ActionId)
			require.NoError(t, err)
			assert.Equal(t, rootCR.Name, childCR.Labels["flyte.org/parent-task-action"])
			assert.Equal(t, "true", childCR.Labels[labelParentLabelEnabled])

			grandchild := newAction()
			grandchild.ActionId.Name = "grandchild"
			grandchild.ParentActionName = &child.ActionId.Name
			require.NoError(t, c.Enqueue(ctx, grandchild, nil))
			grandchildCR, err := c.GetTaskAction(ctx, grandchild.ActionId)
			require.NoError(t, err)
			assert.Equal(t, childCR.Name, grandchildCR.Labels["flyte.org/parent-task-action"])
		})
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
