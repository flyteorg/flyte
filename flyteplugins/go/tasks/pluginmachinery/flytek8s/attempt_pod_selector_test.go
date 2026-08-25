package flytek8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/labels"

	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	pluginsCoreMock "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
)

// completeAttemptLabels mirrors what NewTaskExecutionMetadata injects into every task's
// labels and what every plugin then merges into the Pod templates it builds.
func completeAttemptLabels() map[string]string {
	return map[string]string{
		ManagedLabelKey: ManagedLabelValue,
		RunLabel:        "run-abc",
		ActionLabel:     "a0",
		AttemptLabel:    "1",
		TaskNameLabel:   "train",
	}
}

func metadataWithLabels(podLabels map[string]string) pluginsCore.TaskExecutionMetadata {
	meta := &pluginsCoreMock.TaskExecutionMetadata{}
	meta.EXPECT().GetLabels().Return(podLabels)
	return meta
}

func TestAttemptPodSelector(t *testing.T) {
	t.Run("matches a pod carrying the attempt's labels", func(t *testing.T) {
		selector := AttemptPodSelector(metadataWithLabels(completeAttemptLabels()))
		require.NotNil(t, selector)

		assert.True(t, selector.Matches(labels.Set(completeAttemptLabels())))
	})

	t.Run("ignores labels it does not select on", func(t *testing.T) {
		selector := AttemptPodSelector(metadataWithLabels(completeAttemptLabels()))
		require.NotNil(t, selector)

		// A worker pod carries the operator's own labels alongside the attempt's; they
		// must not stop it matching.
		podLabels := completeAttemptLabels()
		podLabels["ray.io/cluster"] = "cluster-abcde"
		podLabels["ray.io/node-type"] = "worker"
		assert.True(t, selector.Matches(labels.Set(podLabels)))
	})

	t.Run("does not match another attempt of the same action", func(t *testing.T) {
		selector := AttemptPodSelector(metadataWithLabels(completeAttemptLabels()))
		require.NotNil(t, selector)

		podLabels := completeAttemptLabels()
		podLabels[AttemptLabel] = "2"
		assert.False(t, selector.Matches(labels.Set(podLabels)))
	})

	t.Run("does not match another action of the same run", func(t *testing.T) {
		selector := AttemptPodSelector(metadataWithLabels(completeAttemptLabels()))
		require.NotNil(t, selector)

		podLabels := completeAttemptLabels()
		podLabels[ActionLabel] = "a1"
		assert.False(t, selector.Matches(labels.Set(podLabels)))
	})

	t.Run("does not match an unmanaged pod", func(t *testing.T) {
		selector := AttemptPodSelector(metadataWithLabels(completeAttemptLabels()))
		require.NotNil(t, selector)

		podLabels := completeAttemptLabels()
		delete(podLabels, ManagedLabelKey)
		assert.False(t, selector.Matches(labels.Set(podLabels)))
	})

	// A run or action name that sanitizes to nothing is left off the pod rather than
	// stamped blank, so the attempt cannot be identified and no selector may be built.
	// Selecting on what is left would reach another action's pods.
	for _, missing := range []string{ManagedLabelKey, RunLabel, ActionLabel, AttemptLabel} {
		t.Run("refuses to select without "+missing, func(t *testing.T) {
			podLabels := completeAttemptLabels()
			delete(podLabels, missing)
			assert.Nil(t, AttemptPodSelector(metadataWithLabels(podLabels)))
		})

		t.Run("refuses to select on a blank "+missing, func(t *testing.T) {
			podLabels := completeAttemptLabels()
			podLabels[missing] = ""
			assert.Nil(t, AttemptPodSelector(metadataWithLabels(podLabels)))
		})
	}

	t.Run("refuses to select without metadata", func(t *testing.T) {
		assert.Nil(t, AttemptPodSelector(nil))
		assert.Nil(t, AttemptPodSelector(metadataWithLabels(nil)))
	})
}

func TestPreservedPodLabels(t *testing.T) {
	t.Run("carries the identity the selector looks for", func(t *testing.T) {
		preserved := PreservedPodLabels(metadataWithLabels(completeAttemptLabels()))
		selector := AttemptPodSelector(metadataWithLabels(completeAttemptLabels()))
		require.NotNil(t, selector)

		// The point of the helper: what a plugin re-applies last is exactly what the
		// selector goes looking for, so the two can never drift apart.
		assert.True(t, selector.Matches(labels.Set(preserved)))
	})

	t.Run("drops the labels a user supplied over the identity", func(t *testing.T) {
		preserved := PreservedPodLabels(metadataWithLabels(completeAttemptLabels()))

		// This is what a plugin's union does: the user's labels first, the preserved set
		// applied over them last.
		podLabels := map[string]string{
			RunLabel:        "a-run-the-user-made-up",
			ActionLabel:     "not-this-action",
			AttemptLabel:    "99",
			ManagedLabelKey: "false",
			"their-label":   "kept",
		}
		for key, value := range preserved {
			podLabels[key] = value
		}

		selector := AttemptPodSelector(metadataWithLabels(completeAttemptLabels()))
		require.NotNil(t, selector)
		assert.True(t, selector.Matches(labels.Set(podLabels)))
		assert.Equal(t, "kept", podLabels["their-label"])
	})

	t.Run("always carries the managed label", func(t *testing.T) {
		// Even when the attempt cannot be identified, the label the Pod cache selects on
		// has to survive, or the executor cannot see its own pod at all.
		incomplete := completeAttemptLabels()
		delete(incomplete, RunLabel)

		preserved := PreservedPodLabels(metadataWithLabels(incomplete))
		assert.Equal(t, ManagedLabelValue, preserved[ManagedLabelKey])
		assert.NotContains(t, preserved, ActionLabel)

		assert.Equal(t, map[string]string{ManagedLabelKey: ManagedLabelValue}, PreservedPodLabels(nil))
	})
}

func TestAttemptIdentityLabels(t *testing.T) {
	t.Run("returns only the identity, not every execution label", func(t *testing.T) {
		podLabels := completeAttemptLabels()
		podLabels["unrelated"] = "value"

		identity := AttemptIdentityLabels(metadataWithLabels(podLabels))

		assert.Equal(t, map[string]string{
			ManagedLabelKey: ManagedLabelValue,
			RunLabel:        "run-abc",
			ActionLabel:     "a0",
			AttemptLabel:    "1",
		}, identity)
	})

	t.Run("refuses an incomplete identity", func(t *testing.T) {
		podLabels := completeAttemptLabels()
		delete(podLabels, AttemptLabel)
		assert.Nil(t, AttemptIdentityLabels(metadataWithLabels(podLabels)))
	})
}
