package controller

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

// An in-place restart replaces the failure with a Queued phase, so the failure is never
// reported and the reasons it carried are the only account of it the user ever gets. For a
// GPU fault that is the line naming which worker the hardware failed on.
func TestQueuedForInPlaceRestart(t *testing.T) {
	restartedAt := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)
	faultedAt := restartedAt.Add(-2 * time.Minute)

	t.Run("carries the failed attempt's reasons", func(t *testing.T) {
		failed := pluginsCore.PhaseInfoRetryableFailure("TaskFailedWithError", "the job failed",
			&pluginsCore.TaskInfo{
				OccurredAt: &faultedAt,
				AdditionalReasons: []pluginsCore.ReasonInfo{
					{Reason: "GPU fault recorded on pod job-worker-3", OccurredAt: &faultedAt},
				},
			})

		got := queuedForInPlaceRestart(failed, restartedAt)

		assert.Equal(t, pluginsCore.PhaseQueued, got.Phase())
		require.NotNil(t, got.Info())
		require.Len(t, got.Info().AdditionalReasons, 1)
		assert.Equal(t, "GPU fault recorded on pod job-worker-3", got.Info().AdditionalReasons[0].Reason)

		// And they reach the user, which is the point of carrying them.
		events := toClusterEvents(got, nil)
		messages := make([]string, 0, len(events))
		for _, e := range events {
			messages = append(messages, e.GetMessage())
		}
		assert.Contains(t, messages, "GPU fault recorded on pod job-worker-3")
	})

	t.Run("leaves the dead pod's own task info behind", func(t *testing.T) {
		// Logs and custom info describe the pod that just went away, not the one being
		// queued, so carrying them would point the user at a pod that no longer exists.
		failed := pluginsCore.PhaseInfoRetryableFailure("TaskFailedWithError", "the job failed",
			&pluginsCore.TaskInfo{
				OccurredAt: &faultedAt,
				Logs:       []*core.TaskLog{{Uri: "http://logs/the-dead-pod", Name: "dead"}},
			})

		got := queuedForInPlaceRestart(failed, restartedAt)

		require.NotNil(t, got.Info())
		assert.Empty(t, got.Info().Logs)
		require.NotNil(t, got.Info().OccurredAt)
		assert.Equal(t, restartedAt, *got.Info().OccurredAt)
	})

	t.Run("survives a failure that carried no info at all", func(t *testing.T) {
		got := queuedForInPlaceRestart(pluginsCore.PhaseInfoRetryableFailure("Interrupted", "gone", nil), restartedAt)

		assert.Equal(t, pluginsCore.PhaseQueued, got.Phase())
		require.NotNil(t, got.Info())
		assert.Empty(t, got.Info().AdditionalReasons)
	})
}
