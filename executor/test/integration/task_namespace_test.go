package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	executorv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	executorconfig "github.com/flyteorg/flyte/v2/executor/pkg/config"
	"github.com/flyteorg/flyte/v2/executor/pkg/plugin"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

func allowTaskNamespaces(t *testing.T, namespaces ...string) {
	t.Helper()
	cfg := executorconfig.GetConfig()
	original := cfg.AllowedTaskNamespaces
	cfg.AllowedTaskNamespaces = namespaces
	t.Cleanup(func() { cfg.AllowedTaskNamespaces = original })
}

func namespaceTask(namespace, prefix string) *actions.Action {
	runID := &common.RunIdentifier{Project: "test-project", Domain: "development", Name: uniqueRunName(prefix)}
	action := newRootTaskAction(runID)
	action.GetTask().GetSpec().GetTaskTemplate().Config = map[string]string{flytek8s.TaskNamespaceConfigKey: namespace}
	return action
}

func reconcileNamespaceTask(t *testing.T, action *actions.Action) types.NamespacedName {
	t.Helper()
	require.NoError(t, actionsClient.Enqueue(ctx, action, nil))
	key := types.NamespacedName{Namespace: "flyte", Name: action.ActionId.Run.Name + "-a0"}
	for i := 0; i < 2; i++ {
		_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: key})
		require.NoError(t, err)
	}
	return key
}

func TestTaskDefinitionNamespaceCreatesAndCleansUpPod(t *testing.T) {
	const namespace = "task-ns-workers"
	allowTaskNamespaces(t, namespace)
	require.NoError(t, k8sClient.Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}))
	action := namespaceTask(namespace, "task-ns")
	key := reconcileNamespaceTask(t, action)
	podKey := types.NamespacedName{Namespace: namespace, Name: key.Name + "-0"}
	pod := &corev1.Pod{}
	require.NoError(t, k8sClient.Get(ctx, podKey, pod))
	require.Empty(t, pod.OwnerReferences, "a pod cannot reference a TaskAction in another namespace")
	require.Equal(t, key.Namespace, pod.Annotations[flytek8s.TaskActionNamespaceAnnotation])
	require.Equal(t, key.Name, pod.Annotations[flytek8s.TaskActionNameAnnotation])
	// No workload was accidentally launched in the control namespace.
	require.True(t, apierrors.IsNotFound(k8sClient.Get(ctx, types.NamespacedName{Namespace: key.Namespace, Name: podKey.Name}, &corev1.Pod{})))

	// Removing permission for new launches must not strand an accepted workload.
	executorconfig.GetConfig().AllowedTaskNamespaces = nil
	// Simulate a completed task's empty output so the real pod plugin can finish.
	ta := &executorv1.TaskAction{}
	require.NoError(t, k8sClient.Get(ctx, key, ta))
	outputPrefix, err := plugin.ComputeActionOutputPath(ctx, key.Namespace, key.Name, ta.Spec.RunOutputBase, ta.Spec.ActionName, 1)
	require.NoError(t, err)
	require.NoError(t, dataStore.WriteProtobuf(ctx, storage.DataReference(string(outputPrefix)+"/outputs.pb"), storage.Options{}, &core.LiteralMap{}))
	// A terminal workload is retained for logs until its TaskAction is deleted.
	pod.Status.Phase = corev1.PodSucceeded
	require.NoError(t, k8sClient.Status().Update(ctx, pod))
	require.Eventually(t, func() bool {
		_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: key})
		if err != nil {
			return false
		}
		ta := &executorv1.TaskAction{}
		return k8sClient.Get(ctx, key, ta) == nil && ta.Status.PluginPhase == pluginsCore.PhaseSuccess.String()
	}, 10*time.Second, 100*time.Millisecond)
	require.NoError(t, k8sClient.Get(ctx, key, ta))
	require.NoError(t, k8sClient.Delete(ctx, ta))
	_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: key})
	require.NoError(t, err)
	require.True(t, apierrors.IsNotFound(k8sClient.Get(ctx, podKey, &corev1.Pod{})), "deletion must clean up the cross-namespace pod")
}

func TestTaskNamespaceRejectsInvalidAndUnauthorizedTargets(t *testing.T) {
	for _, namespace := range []string{"Bad_namespace", "task-ns-not-allowed"} {
		t.Run(namespace, func(t *testing.T) {
			allowTaskNamespaces(t)
			action := namespaceTask(namespace, "ns-bad")
			key := reconcileNamespaceTask(t, action)
			ta := &executorv1.TaskAction{}
			require.NoError(t, k8sClient.Get(ctx, key, ta))
			require.Empty(t, ta.Finalizers)
			require.Equal(t, metav1.ConditionTrue, ta.Status.Conditions[0].Status)
			require.Equal(t, string(executorv1.ConditionReasonInvalidSpec), ta.Status.Conditions[0].Reason)
		})
	}
}

func TestTaskNamespaceMissingFailsWithoutCreatingNamespace(t *testing.T) {
	const namespace = "task-ns-missing"
	allowTaskNamespaces(t, namespace)
	action := namespaceTask(namespace, "ns-missing")
	key := reconcileNamespaceTask(t, action)
	ta := &executorv1.TaskAction{}
	require.NoError(t, k8sClient.Get(ctx, key, ta))
	require.NotNil(t, ta.Status.ErrorState)
	require.Equal(t, "NamespaceNotFound", ta.Status.ErrorState.Code)
	require.True(t, apierrors.IsNotFound(k8sClient.Get(ctx, types.NamespacedName{Name: namespace}, &corev1.Namespace{})))
}

func TestTaskNamespaceHonorsGPUResourceQuota(t *testing.T) {
	const namespace = "task-ns-quota"
	allowTaskNamespaces(t, namespace)
	require.NoError(t, k8sClient.Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}))
	gpu := corev1.ResourceName("requests.nvidia.com/gpu")
	quota := &corev1.ResourceQuota{ObjectMeta: metav1.ObjectMeta{Name: "gpu-budget", Namespace: namespace},
		Spec: corev1.ResourceQuotaSpec{Hard: corev1.ResourceList{gpu: resource.MustParse("1")}}}
	require.NoError(t, k8sClient.Create(ctx, quota))
	// envtest has no quota controller; initialize its observed quota. The real
	// API-server admission evaluator accounts for Pod requests synchronously.
	quota.Status.Hard = quota.Spec.Hard.DeepCopy()
	quota.Status.Used = corev1.ResourceList{gpu: resource.MustParse("0")}
	require.NoError(t, k8sClient.Status().Update(ctx, quota))
	newGPUAction := func() *actions.Action {
		action := namespaceTask(namespace, "ns-gpu")
		action.GetTask().Spec.TaskTemplate.GetContainer().Resources = &core.Resources{
			Requests: []*core.Resources_ResourceEntry{{Name: core.Resources_GPU, Value: "1"}},
			Limits:   []*core.Resources_ResourceEntry{{Name: core.Resources_GPU, Value: "1"}},
		}
		return action
	}
	first := reconcileNamespaceTask(t, newGPUAction())
	pod := &corev1.Pod{}
	require.NoError(t, k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: first.Name + "-0"}, pod))
	quantity := pod.Spec.Containers[0].Resources.Limits[corev1.ResourceName("nvidia.com/gpu")]
	require.Equal(t, "1", quantity.String())
	second := reconcileNamespaceTask(t, newGPUAction())
	require.True(t, apierrors.IsNotFound(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: second.Name + "-0"}, &corev1.Pod{})))
	ta := &executorv1.TaskAction{}
	require.NoError(t, k8sClient.Get(ctx, second, ta))
	require.Contains(t, fmt.Sprint(ta.Status.Conditions), "exceeded quota")
}

func TestChildTaskSelectsNamespaceIndependently(t *testing.T) {
	const rootNamespace = "task-ns-parent"
	const childNamespace = "task-ns-child"
	allowTaskNamespaces(t, rootNamespace, childNamespace)
	for _, namespace := range []string{rootNamespace, childNamespace} {
		require.NoError(t, k8sClient.Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}))
	}
	root := namespaceTask(rootNamespace, "ns-parent")
	rootKey := reconcileNamespaceTask(t, root)
	child := newRootTaskAction(root.ActionId.Run)
	child.ActionId.Name = "child"
	child.ParentActionName = &root.ActionId.Run.Name
	child.GetTask().Spec.TaskTemplate.Config = map[string]string{flytek8s.TaskNamespaceConfigKey: childNamespace}
	require.NoError(t, actionsClient.Enqueue(ctx, child, nil))
	childKey := types.NamespacedName{Namespace: "flyte", Name: root.ActionId.Run.Name + "-child"}
	for i := 0; i < 2; i++ {
		_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: childKey})
		require.NoError(t, err)
	}
	require.NoError(t, k8sClient.Get(ctx, types.NamespacedName{Namespace: rootNamespace, Name: rootKey.Name + "-0"}, &corev1.Pod{}))
	require.NoError(t, k8sClient.Get(ctx, types.NamespacedName{Namespace: childNamespace, Name: childKey.Name + "-0"}, &corev1.Pod{}))
	childAction := &executorv1.TaskAction{}
	require.NoError(t, k8sClient.Get(ctx, childKey, childAction))
	require.Len(t, childAction.OwnerReferences, 1)
	require.Equal(t, rootKey.Name, childAction.OwnerReferences[0].Name, "TaskAction ownership stays in the control namespace")
}
