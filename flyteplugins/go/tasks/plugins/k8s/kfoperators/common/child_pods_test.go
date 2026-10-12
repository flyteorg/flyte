package common

import (
	"context"
	"testing"

	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
)

// AttemptExecutionLabels mirrors what NewTaskExecutionMetadata stamps on every task and
// what ToReplicaSpec then merges onto every replica pod template.
func AttemptExecutionLabels() map[string]string {
	return map[string]string{
		"label-key":              "label-value",
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

func TestChildPods(t *testing.T) {
	job := &kubeflowv1.PyTorchJob{
		ObjectMeta: metav1.ObjectMeta{Namespace: "test-namespace", Name: "job3"},
	}

	t.Run("selects on the attempt and the job", func(t *testing.T) {
		selector, err := ChildPods(context.TODO(), attemptMetadata(AttemptExecutionLabels()), job)

		require.NoError(t, err)
		require.NotNil(t, selector)

		podLabels := AttemptExecutionLabels()
		podLabels[kubeflowv1.JobNameLabel] = "job3"
		podLabels[kubeflowv1.ReplicaTypeLabel] = "worker"
		podLabels[kubeflowv1.ReplicaIndexLabel] = "0"
		assert.True(t, selector.Matches(labels.Set(podLabels)))

		// Another job in the same namespace is not this one.
		podLabels[kubeflowv1.JobNameLabel] = "job4"
		assert.False(t, selector.Matches(labels.Set(podLabels)))
	})

	t.Run("does not select another attempt of the same action", func(t *testing.T) {
		selector, err := ChildPods(context.TODO(), attemptMetadata(AttemptExecutionLabels()), job)
		require.NoError(t, err)
		require.NotNil(t, selector)

		podLabels := AttemptExecutionLabels()
		podLabels[kubeflowv1.JobNameLabel] = "job3"
		podLabels[flytek8s.AttemptLabel] = "2"
		assert.False(t, selector.Matches(labels.Set(podLabels)))
	})

	t.Run("declines when the attempt cannot be identified", func(t *testing.T) {
		executionLabels := AttemptExecutionLabels()
		delete(executionLabels, flytek8s.ActionLabel)

		selector, err := ChildPods(context.TODO(), attemptMetadata(executionLabels), job)

		require.NoError(t, err)
		assert.Nil(t, selector)
	})

	t.Run("rejects a missing resource", func(t *testing.T) {
		_, err := ChildPods(context.TODO(), attemptMetadata(AttemptExecutionLabels()), nil)
		assert.Error(t, err)
	})
}
