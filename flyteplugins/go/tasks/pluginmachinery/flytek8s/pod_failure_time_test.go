package flytek8s

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func terminatedAt(at time.Time) v1.ContainerStatus {
	return v1.ContainerStatus{
		State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{FinishedAt: metav1.NewTime(at)}},
	}
}

func runningSince(at time.Time) v1.ContainerStatus {
	return v1.ContainerStatus{
		State: v1.ContainerState{Running: &v1.ContainerStateRunning{StartedAt: metav1.NewTime(at)}},
	}
}

func TestPodFailureTime(t *testing.T) {
	occurredAt := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)

	t.Run("prefers the latest container termination", func(t *testing.T) {
		first := occurredAt.Add(-10 * time.Minute)
		last := occurredAt.Add(-2 * time.Minute)
		pod := &v1.Pod{Status: v1.PodStatus{ContainerStatuses: []v1.ContainerStatus{
			terminatedAt(first),
			terminatedAt(last),
		}}}
		assert.Equal(t, last, PodFailureTime(pod, occurredAt))
	})

	t.Run("ignores a container that has not terminated", func(t *testing.T) {
		died := occurredAt.Add(-2 * time.Minute)
		pod := &v1.Pod{Status: v1.PodStatus{ContainerStatuses: []v1.ContainerStatus{
			runningSince(occurredAt.Add(-6 * time.Hour)),
			terminatedAt(died),
		}}}
		assert.Equal(t, died, PodFailureTime(pod, occurredAt))
	})

	t.Run("ignores a termination with no finish time", func(t *testing.T) {
		pod := &v1.Pod{Status: v1.PodStatus{ContainerStatuses: []v1.ContainerStatus{
			{State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{}}},
		}}}
		assert.Equal(t, occurredAt, PodFailureTime(pod, occurredAt))
	})

	t.Run("ignores init containers", func(t *testing.T) {
		// This is where it parts company with GetLastTransitionOccurredAt. An init
		// container finished before the work started, and a native sidecar declared among
		// the init containers is reaped after everything else. Anchoring on either would
		// name a moment that has nothing to do with when the work died.
		initFinished := occurredAt.Add(-3 * time.Hour)
		pod := &v1.Pod{Status: v1.PodStatus{
			InitContainerStatuses: []v1.ContainerStatus{terminatedAt(initFinished)},
			ContainerStatuses:     []v1.ContainerStatus{runningSince(occurredAt.Add(-3 * time.Hour))},
		}}
		assert.Equal(t, occurredAt, PodFailureTime(pod, occurredAt))
		assert.NotEqual(t, initFinished, PodFailureTime(pod, occurredAt))
	})

	t.Run("falls back to the deletion timestamp", func(t *testing.T) {
		// What an eviction leaves behind: nothing terminated, but the pod is on its way
		// out and the API server stamped when.
		deletedAt := occurredAt.Add(-time.Minute)
		deletion := metav1.NewTime(deletedAt)
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{DeletionTimestamp: &deletion},
			Status:     v1.PodStatus{ContainerStatuses: []v1.ContainerStatus{runningSince(occurredAt.Add(-6 * time.Hour))}},
		}
		assert.Equal(t, deletedAt, PodFailureTime(pod, occurredAt))
	})

	t.Run("prefers a termination over the deletion timestamp", func(t *testing.T) {
		died := occurredAt.Add(-5 * time.Minute)
		deletion := metav1.NewTime(occurredAt.Add(-time.Minute))
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{DeletionTimestamp: &deletion},
			Status:     v1.PodStatus{ContainerStatuses: []v1.ContainerStatus{terminatedAt(died)}},
		}
		assert.Equal(t, died, PodFailureTime(pod, occurredAt))
	})

	t.Run("falls back to the time the caller offered, then to now", func(t *testing.T) {
		assert.Equal(t, occurredAt, PodFailureTime(&v1.Pod{}, occurredAt))
		assert.WithinDuration(t, time.Now(), PodFailureTime(&v1.Pod{}, time.Time{}), time.Minute)
	})
}

// The anchor exists because GetLastTransitionOccurredAt cannot serve as one. For a pod
// that failed while its containers were still running it reports the time the container
// started, so a long-running task would be anchored hours before anything went wrong.
func TestPodFailureTimeDoesNotAnchorOnAStartTime(t *testing.T) {
	startedAt := time.Now().Add(-6 * time.Hour)
	evictedAt := time.Now().Add(-2 * time.Minute)

	deletion := metav1.NewTime(evictedAt)
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{DeletionTimestamp: &deletion},
		Status:     v1.PodStatus{ContainerStatuses: []v1.ContainerStatus{runningSince(startedAt)}},
	}

	assert.Equal(t, startedAt.Unix(), GetLastTransitionOccurredAt(pod).Unix())
	assert.Equal(t, evictedAt, PodFailureTime(pod, GetLastTransitionOccurredAt(pod).Time))
}
