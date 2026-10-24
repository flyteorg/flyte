package k8s

import (
	"context"
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

const projectDomainTemplate = "{{ project }}-{{ domain }}"

// newMappedClient returns an ActionsClient that maps runs to "<project>-<domain>",
// backed by a fake k8s client holding objs.
func newMappedClient(t *testing.T, createNamespaces bool, objs ...client.Object) *ActionsClient {
	return newMappedClientWithFuncs(t, createNamespaces, interceptor.Funcs{}, objs...)
}

// newMappedClientWithFuncs is newMappedClient with k8s calls intercepted by funcs.
func newMappedClientWithFuncs(t *testing.T, createNamespaces bool, funcs interceptor.Funcs, objs ...client.Object) *ActionsClient {
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, executorv1.AddToScheme(scheme))
	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objs...).
		WithStatusSubresource(&executorv1.TaskAction{}).
		WithInterceptorFuncs(funcs).
		Build()
	c := &ActionsClient{
		recordedFilter: testFilter(),
		namespace:      "flyte",
		k8sClient:      k8sClient,
	}
	WithNamespaceMapping(projectDomainTemplate, createNamespaces)(c)
	return c
}

func mappedRunID(project string) *common.RunIdentifier {
	return &common.RunIdentifier{Project: project, Domain: "development", Name: "run1"}
}

func newTask(runID *common.RunIdentifier, name string, parent *string) *actions.Action {
	return &actions.Action{
		ActionId:         &common.ActionIdentifier{Run: runID, Name: name},
		ParentActionName: parent,
		Spec: &actions.Action_Task{
			Task: &workflow.TaskAction{
				Spec: &task.TaskSpec{TaskTemplate: &core.TaskTemplate{Type: "python"}},
			},
		},
	}
}

func namespaceObj(name string) *corev1.Namespace {
	return &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}
}

// legacyRoot is the root TaskAction of a run started before the namespace
// template was set, so it is in the default namespace.
func legacyRoot(runID *common.RunIdentifier) *executorv1.TaskAction {
	return &executorv1.TaskAction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      runID.Name + "-a0",
			Namespace: "flyte",
			Labels: map[string]string{
				"flyte.org/project": runID.Project,
				"flyte.org/domain":  runID.Domain,
				"flyte.org/run":     runID.Name,
			},
		},
		Spec: executorv1.TaskActionSpec{ActionName: runID.Name},
	}
}

func TestEnqueue_RootGoesToMappedNamespace(t *testing.T) {
	ctx := context.Background()
	c := newMappedClient(t, true)
	runID := mappedRunID("flytesnacks")

	require.NoError(t, c.Enqueue(ctx, newTask(runID, runID.Name, nil), nil))

	root, err := c.GetTaskAction(ctx, &common.ActionIdentifier{Run: runID, Name: runID.Name})
	require.NoError(t, err)
	assert.Equal(t, "flytesnacks-development", root.Namespace)

	// The mapped namespace was created for the run.
	require.NoError(t, c.k8sClient.Get(ctx, client.ObjectKey{Name: "flytesnacks-development"}, &corev1.Namespace{}))
}

func TestEnqueue_ChildGoesToParentNamespace(t *testing.T) {
	ctx := context.Background()
	c := newMappedClient(t, true)
	runID := mappedRunID("flytesnacks")
	require.NoError(t, c.Enqueue(ctx, newTask(runID, runID.Name, nil), nil))

	parent := runID.Name
	require.NoError(t, c.Enqueue(ctx, newTask(runID, "a1", &parent), nil))

	child, err := c.GetTaskAction(ctx, &common.ActionIdentifier{Run: runID, Name: "a1"})
	require.NoError(t, err)
	assert.Equal(t, "flytesnacks-development", child.Namespace)
	require.Len(t, child.OwnerReferences, 1)
	assert.Equal(t, "run1-a0", child.OwnerReferences[0].Name)

	children, err := c.ListChildActions(ctx, &common.ActionIdentifier{Run: runID, Name: runID.Name})
	require.NoError(t, err)
	assert.Len(t, children, 2) // the parent and its child
}

// A run started before the namespace template was set keeps its TaskActions in
// the default namespace: children follow the parent there, and lookups find them.
func TestEnqueue_RunStartedBeforeMappingStaysInDefaultNamespace(t *testing.T) {
	ctx := context.Background()
	runID := mappedRunID("flytesnacks")
	c := newMappedClient(t, true, legacyRoot(runID))

	parent := runID.Name
	require.NoError(t, c.Enqueue(ctx, newTask(runID, "a1", &parent), nil))

	child, err := c.GetTaskAction(ctx, &common.ActionIdentifier{Run: runID, Name: "a1"})
	require.NoError(t, err)
	assert.Equal(t, "flyte", child.Namespace)

	children, err := c.ListChildActions(ctx, &common.ActionIdentifier{Run: runID, Name: runID.Name})
	require.NoError(t, err)
	assert.Len(t, children, 2)

	require.NoError(t, c.AbortAction(ctx, &common.ActionIdentifier{Run: runID, Name: runID.Name}, nil))
	err = c.k8sClient.Get(ctx, client.ObjectKey{Name: "run1-a0", Namespace: "flyte"}, &executorv1.TaskAction{})
	assert.True(t, apierrors.IsNotFound(err), "abort must delete the root in the default namespace")
}

// Resubmitting a run started before the namespace template was set, as the
// scheduler does when it fires a run again, must not start a second root.
func TestEnqueue_ResubmittedLegacyRootIsNotLaunchedAgain(t *testing.T) {
	ctx := context.Background()
	runID := mappedRunID("flytesnacks")
	c := newMappedClient(t, true, legacyRoot(runID))

	err := c.Enqueue(ctx, newTask(runID, runID.Name, nil), nil)
	assert.Equal(t, connect.CodeAlreadyExists, connect.CodeOf(err))

	all := &executorv1.TaskActionList{}
	require.NoError(t, c.k8sClient.List(ctx, all))
	require.Len(t, all.Items, 1)
	assert.Equal(t, "flyte", all.Items[0].Namespace)
}

// TaskAction names hold only the run name, so the default namespace can hold a
// same-named action of another project's run started before the mapping. It must
// not be read, aborted or mistaken for this run's root.
func TestLegacyActionOfAnotherProjectIsIgnored(t *testing.T) {
	ctx := context.Background()
	runA, runB := mappedRunID("proj-a"), mappedRunID("proj-b")
	c := newMappedClient(t, true, legacyRoot(runA))
	rootB := &common.ActionIdentifier{Run: runB, Name: runB.Name}

	_, err := c.GetTaskAction(ctx, rootB)
	assert.True(t, apierrors.IsNotFound(err), "got %v", err)
	assert.Error(t, c.AbortAction(ctx, rootB, nil))
	require.NoError(t, c.k8sClient.Get(ctx, client.ObjectKeyFromObject(legacyRoot(runA)), &executorv1.TaskAction{}),
		"project A's action must survive an abort of project B's run")

	require.NoError(t, c.Enqueue(ctx, newTask(runB, runB.Name, nil), nil))
	b, err := c.GetTaskAction(ctx, rootB)
	require.NoError(t, err)
	assert.Equal(t, "proj-b-development", b.Namespace)
}

func TestEnqueue_InvalidMappedNamespace(t *testing.T) {
	ctx := context.Background()
	c := newMappedClient(t, true)
	runID := mappedRunID("Flyte_Snacks")

	err := c.Enqueue(ctx, newTask(runID, runID.Name, nil), nil)
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

func TestEnqueue_RequireExistingNamespaces(t *testing.T) {
	ctx := context.Background()

	t.Run("missing namespace fails the run", func(t *testing.T) {
		c := newMappedClient(t, false)
		runID := mappedRunID("flytesnacks")

		err := c.Enqueue(ctx, newTask(runID, runID.Name, nil), nil)
		require.Error(t, err)
		assert.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))
		assert.ErrorContains(t, err, `namespace "flytesnacks-development" does not exist`)

		// Nothing was created on the run's behalf.
		err = c.k8sClient.Get(ctx, client.ObjectKey{Name: "flytesnacks-development"}, &corev1.Namespace{})
		assert.True(t, apierrors.IsNotFound(err))
	})

	t.Run("existing namespace is used", func(t *testing.T) {
		c := newMappedClient(t, false, namespaceObj("flytesnacks-development"))
		runID := mappedRunID("flytesnacks")

		require.NoError(t, c.Enqueue(ctx, newTask(runID, runID.Name, nil), nil))
		root, err := c.GetTaskAction(ctx, &common.ActionIdentifier{Run: runID, Name: runID.Name})
		require.NoError(t, err)
		assert.Equal(t, "flytesnacks-development", root.Namespace)
	})
}

// A namespace is checked once and then remembered, until creating a TaskAction in
// it reports that it no longer exists.
func TestEnqueue_RemembersExistingNamespaces(t *testing.T) {
	ctx := context.Background()
	namespaceGets := 0
	failNextCreate := false
	c := newMappedClientWithFuncs(t, false, interceptor.Funcs{
		Get: func(ctx context.Context, cl client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			if _, ok := obj.(*corev1.Namespace); ok {
				namespaceGets++
			}
			return cl.Get(ctx, key, obj, opts...)
		},
		Create: func(ctx context.Context, cl client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
			if failNextCreate {
				failNextCreate = false
				return apierrors.NewNotFound(schema.GroupResource{Resource: "namespaces"}, obj.GetNamespace())
			}
			return cl.Create(ctx, obj, opts...)
		},
	}, namespaceObj("flytesnacks-development"))

	enqueueRoot := func(runName string) error {
		runID := &common.RunIdentifier{Project: "flytesnacks", Domain: "development", Name: runName}
		return c.Enqueue(ctx, newTask(runID, runName, nil), nil)
	}

	require.NoError(t, enqueueRoot("run1"))
	require.NoError(t, enqueueRoot("run2"))
	assert.Equal(t, 1, namespaceGets, "the namespace should be checked once and then remembered")

	// A namespace deleted after it was remembered fails the run, as nothing retries it.
	failNextCreate = true
	assert.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(enqueueRoot("run3")))
	require.NoError(t, enqueueRoot("run4"))
	assert.Equal(t, 2, namespaceGets, "a namespace reported missing should be checked again")
}

// Runs with the same name in two projects get their own TaskActions.
func TestEnqueue_SameRunNameInTwoProjects(t *testing.T) {
	ctx := context.Background()
	c := newMappedClient(t, true)
	runA, runB := mappedRunID("proj-a"), mappedRunID("proj-b")

	require.NoError(t, c.Enqueue(ctx, newTask(runA, runA.Name, nil), nil))
	require.NoError(t, c.Enqueue(ctx, newTask(runB, runB.Name, nil), nil))

	a, err := c.GetTaskAction(ctx, &common.ActionIdentifier{Run: runA, Name: runA.Name})
	require.NoError(t, err)
	b, err := c.GetTaskAction(ctx, &common.ActionIdentifier{Run: runB, Name: runB.Name})
	require.NoError(t, err)
	assert.Equal(t, "proj-a-development", a.Namespace)
	assert.Equal(t, "proj-b-development", b.Namespace)
}

func TestEnqueue_NoTemplateUsesDefaultNamespace(t *testing.T) {
	ctx := context.Background()
	c := newMappedClient(t, true)
	c.namespaceTemplate = ""
	runID := mappedRunID("flytesnacks")

	require.NoError(t, c.Enqueue(ctx, newTask(runID, runID.Name, nil), nil))
	root, err := c.GetTaskAction(ctx, &common.ActionIdentifier{Run: runID, Name: runID.Name})
	require.NoError(t, err)
	assert.Equal(t, "flyte", root.Namespace)
}
