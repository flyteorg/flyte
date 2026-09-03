package ray

import (
	"context"
	"testing"

	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

// attemptExecutionLabels mirrors what NewTaskExecutionMetadata stamps on every task.
func attemptExecutionLabels() map[string]string {
	return map[string]string{
		"label-1":                "val1",
		flytek8s.ManagedLabelKey: flytek8s.ManagedLabelValue,
		flytek8s.RunLabel:        "run-abc",
		flytek8s.ActionLabel:     "a0",
		flytek8s.AttemptLabel:    "1",
	}
}

func attemptMetadata(executionLabels map[string]string) pluginsCore.TaskExecutionMetadata {
	meta := &mocks.TaskExecutionMetadata{}
	meta.EXPECT().GetLabels().Return(executionLabels)
	return meta
}

func rayJobWithCluster(namespace, clusterName string) *rayv1.RayJob {
	return &rayv1.RayJob{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: "job"},
		Status:     rayv1.RayJobStatus{RayClusterName: clusterName},
	}
}

func TestRayChildPods(t *testing.T) {
	t.Run("selects on the attempt and the cluster", func(t *testing.T) {
		selector, err := rayJobResourceHandler{}.ChildPods(
			context.TODO(), attemptMetadata(attemptExecutionLabels()), rayJobWithCluster("test-namespace", "job-abcde"))

		require.NoError(t, err)
		require.NotNil(t, selector)

		podLabels := attemptExecutionLabels()
		podLabels[rayClusterLabelKey] = "job-abcde"
		assert.True(t, selector.Matches(labels.Set(podLabels)))

		// Another cluster in the same namespace is not this attempt's.
		podLabels[rayClusterLabelKey] = "other-fghij"
		assert.False(t, selector.Matches(labels.Set(podLabels)))
	})

	t.Run("falls back to the attempt alone before the cluster is named", func(t *testing.T) {
		// KubeRay appends a random suffix to the cluster name, so it is only knowable
		// from the RayJob's status. Until it is reported the attempt labels stand alone,
		// which is still exact to one attempt of one action.
		selector, err := rayJobResourceHandler{}.ChildPods(
			context.TODO(), attemptMetadata(attemptExecutionLabels()), rayJobWithCluster("test-namespace", ""))

		require.NoError(t, err)
		require.NotNil(t, selector)
		assert.True(t, selector.Matches(labels.Set(attemptExecutionLabels())))

		other := attemptExecutionLabels()
		other[flytek8s.ActionLabel] = "a1"
		assert.False(t, selector.Matches(labels.Set(other)))
	})

	t.Run("declines when the attempt cannot be identified", func(t *testing.T) {
		executionLabels := attemptExecutionLabels()
		delete(executionLabels, flytek8s.RunLabel)

		selector, err := rayJobResourceHandler{}.ChildPods(
			context.TODO(), attemptMetadata(executionLabels), rayJobWithCluster("test-namespace", "job-abcde"))

		require.NoError(t, err)
		assert.Nil(t, selector)
	})

	t.Run("rejects a resource that is not a RayJob", func(t *testing.T) {
		_, err := rayJobResourceHandler{}.ChildPods(
			context.TODO(), attemptMetadata(attemptExecutionLabels()), &corev1.Pod{})

		assert.Error(t, err)
	})
}

// TestRayChildPodsIdentityNotOverridable verifies that a task cannot take its own pods out
// of the framework's reach by setting the identity labels in its k8s_pod metadata. A pod
// whose run, action or attempt label has been overwritten is one the selector cannot find,
// and a GPU fault on it would be silently lost rather than reaching the failure.
func TestRayChildPodsIdentityNotOverridable(t *testing.T) {
	require.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{}))

	rayJobObj := dummyRayCustomObj()
	overrides := &core.K8SPod{
		Metadata: &core.K8SObjectMetadata{
			Labels: map[string]string{
				flytek8s.RunLabel:     "a-run-the-user-made-up",
				flytek8s.ActionLabel:  "not-this-action",
				flytek8s.AttemptLabel: "99",
			},
		},
	}
	rayJobObj.RayCluster.HeadGroupSpec.K8SPod = overrides
	rayJobObj.RayCluster.WorkerGroupSpec[0].K8SPod = overrides

	executionLabels := attemptExecutionLabels()
	taskTemplate := dummyRayTaskTemplate("ray-id", rayJobObj)
	taskCtx := dummyRayTaskContextWithLabels(taskTemplate, resourceRequirements, nil, "", serviceAccount, true, executionLabels)

	resource, err := rayJobResourceHandler{}.BuildResource(context.TODO(), taskCtx)
	require.NoError(t, err)
	rayJob, ok := resource.(*rayv1.RayJob)
	require.True(t, ok)

	rayJob.Status.RayClusterName = "job-abcde"
	selector, err := rayJobResourceHandler{}.ChildPods(context.TODO(), attemptMetadata(executionLabels), rayJob)
	require.NoError(t, err)
	require.NotNil(t, selector)

	for name, templateLabels := range map[string]map[string]string{
		"head":   rayJob.Spec.RayClusterSpec.HeadGroupSpec.Template.GetLabels(),
		"worker": rayJob.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.GetLabels(),
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, executionLabels[flytek8s.RunLabel], templateLabels[flytek8s.RunLabel])
			assert.Equal(t, executionLabels[flytek8s.ActionLabel], templateLabels[flytek8s.ActionLabel])
			assert.Equal(t, executionLabels[flytek8s.AttemptLabel], templateLabels[flytek8s.AttemptLabel])

			podLabels := make(map[string]string, len(templateLabels)+1)
			for k, v := range templateLabels {
				podLabels[k] = v
			}
			podLabels[rayClusterLabelKey] = rayJob.Status.RayClusterName
			assert.True(t, selector.Matches(labels.Set(podLabels)))
		})
	}
}

// TestRayChildPodsMatchTheTemplatesTheyCameFrom is the conformance check between the two
// halves of child pod discovery: the labels this plugin puts on the pod templates KubeRay
// expands, and the selector it hands the framework to find the resulting pods. Label drift
// on either side would otherwise leave a GPU fault on a Ray worker silently unclassified.
func TestRayChildPodsMatchTheTemplatesTheyCameFrom(t *testing.T) {
	executionLabels := attemptExecutionLabels()
	taskTemplate := dummyRayTaskTemplate("ray-id", dummyRayCustomObj())
	taskCtx := dummyRayTaskContextWithLabels(taskTemplate, resourceRequirements, nil, "", serviceAccount, true, executionLabels)

	resource, err := rayJobResourceHandler{}.BuildResource(context.TODO(), taskCtx)
	require.NoError(t, err)
	rayJob, ok := resource.(*rayv1.RayJob)
	require.True(t, ok)

	// KubeRay stamps the cluster label on the pods themselves, so the templates do not
	// carry it and the fixture adds what the operator would.
	rayJob.Status.RayClusterName = "job-abcde"
	selector, err := rayJobResourceHandler{}.ChildPods(context.TODO(), attemptMetadata(executionLabels), rayJob)
	require.NoError(t, err)
	require.NotNil(t, selector)

	templates := map[string]map[string]string{
		"head":      rayJob.Spec.RayClusterSpec.HeadGroupSpec.Template.GetLabels(),
		"worker":    rayJob.Spec.RayClusterSpec.WorkerGroupSpecs[0].Template.GetLabels(),
		"submitter": rayJob.Spec.SubmitterPodTemplate.GetLabels(),
	}
	for name, templateLabels := range templates {
		t.Run(name, func(t *testing.T) {
			require.NotEmpty(t, templateLabels)
			podLabels := make(map[string]string, len(templateLabels)+1)
			for k, v := range templateLabels {
				podLabels[k] = v
			}
			// The submitter is a plain Job pod and carries no cluster label, so it is
			// checked against the attempt half only, which is what excludes it in
			// production. The head and worker pods get the label from KubeRay.
			if name != "submitter" {
				podLabels[rayClusterLabelKey] = rayJob.Status.RayClusterName
				assert.True(t, selector.Matches(labels.Set(podLabels)),
					"the %s pod template's labels %v do not satisfy %s", name, podLabels, selector)
				return
			}
			attemptOnly := flytek8s.AttemptPodSelector(attemptMetadata(executionLabels))
			require.NotNil(t, attemptOnly)
			assert.True(t, attemptOnly.Matches(labels.Set(podLabels)))
		})
	}
}
