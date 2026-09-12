package clustered

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"

	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	coreMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	clusteredpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/plugins"
)

// attemptExecutionLabels mirrors what NewTaskExecutionMetadata stamps on every task and
// what build.go then merges onto every child pod template.
func attemptExecutionLabels() map[string]string {
	return map[string]string{
		"execution-id":           "my-exec",
		"node-id":                "n1",
		flytek8s.ManagedLabelKey: flytek8s.ManagedLabelValue,
		flytek8s.RunLabel:        "run-abc",
		flytek8s.ActionLabel:     "a0",
		flytek8s.AttemptLabel:    "1",
	}
}

func attemptMetadata(executionLabels map[string]string) pluginsCore.TaskExecutionMetadata {
	meta := &coreMocks.TaskExecutionMetadata{}
	meta.EXPECT().GetLabels().Return(executionLabels)
	return meta
}

func TestClusteredChildPods(t *testing.T) {
	jobSet := &jobsetv1alpha2.JobSet{
		ObjectMeta: metav1.ObjectMeta{Namespace: testNS, Name: testJobName},
	}

	t.Run("selects on the attempt and the JobSet", func(t *testing.T) {
		selector, err := clusteredResourceHandler{}.ChildPods(
			context.Background(), attemptMetadata(attemptExecutionLabels()), jobSet)

		require.NoError(t, err)
		require.NotNil(t, selector)

		podLabels := attemptExecutionLabels()
		podLabels[jobsetv1alpha2.JobSetNameKey] = testJobName
		assert.True(t, selector.Matches(labels.Set(podLabels)))

		// Another JobSet in the same namespace is not this one.
		podLabels[jobsetv1alpha2.JobSetNameKey] = "some-other-jobset"
		assert.False(t, selector.Matches(labels.Set(podLabels)))
	})

	t.Run("does not select another attempt of the same action", func(t *testing.T) {
		selector, err := clusteredResourceHandler{}.ChildPods(
			context.Background(), attemptMetadata(attemptExecutionLabels()), jobSet)
		require.NoError(t, err)
		require.NotNil(t, selector)

		podLabels := attemptExecutionLabels()
		podLabels[jobsetv1alpha2.JobSetNameKey] = testJobName
		podLabels[flytek8s.AttemptLabel] = "2"
		assert.False(t, selector.Matches(labels.Set(podLabels)))
	})

	t.Run("declines when the attempt cannot be identified", func(t *testing.T) {
		executionLabels := attemptExecutionLabels()
		delete(executionLabels, flytek8s.AttemptLabel)

		selector, err := clusteredResourceHandler{}.ChildPods(
			context.Background(), attemptMetadata(executionLabels), jobSet)

		require.NoError(t, err)
		assert.Nil(t, selector)
	})

	t.Run("rejects a resource that is not a JobSet", func(t *testing.T) {
		_, err := clusteredResourceHandler{}.ChildPods(
			context.Background(), attemptMetadata(attemptExecutionLabels()), &corev1.Pod{})

		assert.Error(t, err)
	})
}

// TestClusteredChildPodsMatchTheTemplatesTheyCameFrom is the conformance check between the
// two halves of child pod discovery: the labels this plugin puts on the pod templates the
// JobSet controller expands, and the selector it hands the framework to find the resulting
// pods. Label drift on either side would leave a GPU fault on a worker silently
// unclassified, which is the failure mode this whole path exists to prevent.
func TestClusteredChildPodsMatchTheTemplatesTheyCameFrom(t *testing.T) {
	spec := &clusteredpb.ClusteredTaskSpec{
		Replicas:     4,
		NprocPerNode: 8,
		Runtime: &clusteredpb.Runtime{
			Kind: &clusteredpb.Runtime_Torchrun{
				Torchrun: &clusteredpb.TorchRuntime{
					RdzvBackend: clusteredpb.RdzvBackend_STATIC,
				},
			},
		},
		FailurePolicy: &clusteredpb.ClusterFailurePolicy{MaxRestarts: 3},
	}
	executionLabels := attemptExecutionLabels()
	taskCtx := dummyTaskCtxWithLabels(buildTaskTemplate(spec), testJobName, executionLabels)

	obj, err := clusteredResourceHandler{}.BuildResource(context.Background(), taskCtx)
	require.NoError(t, err)
	jobSet, ok := obj.(*jobsetv1alpha2.JobSet)
	require.True(t, ok)

	selector, err := clusteredResourceHandler{}.ChildPods(context.Background(), attemptMetadata(executionLabels), jobSet)
	require.NoError(t, err)
	require.NotNil(t, selector)

	require.NotEmpty(t, jobSet.Spec.ReplicatedJobs)
	for _, replicatedJob := range jobSet.Spec.ReplicatedJobs {
		t.Run(replicatedJob.Name, func(t *testing.T) {
			templateLabels := replicatedJob.Template.Spec.Template.GetLabels()
			require.NotEmpty(t, templateLabels)

			podLabels := make(map[string]string, len(templateLabels)+1)
			for k, v := range templateLabels {
				podLabels[k] = v
			}
			// The JobSet controller stamps its own name on the pods it creates, so the
			// template does not carry it and the fixture adds what the operator would.
			podLabels[jobsetv1alpha2.JobSetNameKey] = jobSet.Name

			assert.True(t, selector.Matches(labels.Set(podLabels)),
				"the %s pod template's labels %v do not satisfy %s", replicatedJob.Name, podLabels, selector)
		})
	}
}
