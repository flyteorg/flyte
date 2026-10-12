package gang

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

const (
	kueueSource      = "kueue"
	preempted        = "Preempted"
	preemptedMessage = "Preempted to accommodate a workload due to prioritization in the ClusterQueue"
)

func TestEvictionExecutionError_Kind(t *testing.T) {
	user := Eviction{Source: kueueSource, Reason: preempted, UserCaused: true, RetriesLeft: 1}.ExecutionError()
	assert.Equal(t, CodeGangEvicted, user.GetCode())
	assert.Equal(t, core.ExecutionError_USER, user.GetKind())

	system := Eviction{Source: kueueSource, Reason: "ClusterQueueStopped"}.ExecutionError()
	assert.Equal(t, CodeGangEvicted, system.GetCode())
	assert.Equal(t, core.ExecutionError_SYSTEM, system.GetKind())
}

func TestEvictionExecutionError_Message(t *testing.T) {
	tests := []struct {
		name     string
		eviction Eviction
		want     string
	}{
		{
			name: "user caused with retries left",
			eviction: Eviction{
				Source: kueueSource, Reason: preempted, Detail: "InClusterQueue",
				Message: "  " + preemptedMessage + "  ", UserCaused: true, RetriesLeft: 3,
			},
			want: "gang evicted by kueue (Preempted, InClusterQueue): " + preemptedMessage +
				"; this uses one of the task's retries (2 left after it)",
		},
		{
			name:     "user caused on the last retry",
			eviction: Eviction{Source: kueueSource, Reason: preempted, UserCaused: true, RetriesLeft: 1},
			want:     "gang evicted by kueue (Preempted); this uses the task's last retry",
		},
		{
			name:     "user caused with no retries left",
			eviction: Eviction{Source: kueueSource, Reason: preempted, UserCaused: true},
			want: "gang evicted by kueue (Preempted); " +
				"this counts against the task's retries and none are left, so the task fails",
		},
		{
			name:     "system caused",
			eviction: Eviction{Source: kueueSource, Reason: "Deactivated", Message: "The workload is deactivated"},
			want: "gang evicted by kueue (Deactivated): The workload is deactivated; " +
				"this does not count against the task's retries",
		},
		{
			name:     "reason unknown",
			eviction: Eviction{Source: kueueSource},
			want:     "gang evicted by kueue; this does not count against the task's retries",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.eviction.ExecutionError().GetMessage())
		})
	}
}
