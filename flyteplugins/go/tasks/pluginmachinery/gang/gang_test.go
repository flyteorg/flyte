package gang

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

const kueueSource = "kueue"

func TestDecide(t *testing.T) {
	tests := []struct {
		name       string
		eviction   Eviction
		policy     Policy
		wantAction Action
		wantKind   core.ExecutionError_ErrorKind
	}{
		{
			name:       "not started holds regardless of policy",
			eviction:   Eviction{Source: kueueSource, Started: false},
			policy:     Policy{AsSystemRetry: true},
			wantAction: Hold,
		},
		{
			name:       "not started holds with user policy",
			eviction:   Eviction{Source: kueueSource, Started: false},
			policy:     Policy{AsSystemRetry: false},
			wantAction: Hold,
		},
		{
			name: "started with system policy is a system retry",
			eviction: Eviction{
				Source:  kueueSource,
				Started: true,
				Reason:  ReasonPreempted,
				Message: "Preempted to accommodate a higher priority Workload",
			},
			policy:     Policy{AsSystemRetry: true},
			wantAction: SystemRetry,
			wantKind:   core.ExecutionError_SYSTEM,
		},
		{
			name:       "started with user policy is a user retry",
			eviction:   Eviction{Source: kueueSource, Started: true},
			policy:     Policy{AsSystemRetry: false},
			wantAction: UserRetry,
			wantKind:   core.ExecutionError_USER,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, err := Decide(tt.eviction, tt.policy)
			assert.Equal(t, tt.wantAction, action)
			if tt.wantAction == Hold {
				assert.Nil(t, err)
				return
			}
			require.NotNil(t, err)
			assert.Equal(t, CodeGangEvicted, err.GetCode())
			assert.Equal(t, tt.wantKind, err.GetKind())
			assert.Contains(t, err.GetMessage(), "gang evicted by kueue")
		})
	}
}

func TestDecide_MessageComposition(t *testing.T) {
	_, err := Decide(Eviction{
		Source:  kueueSource,
		Reason:  ReasonPodsReadyTimeout,
		Message: "  Exceeded the PodsReady timeout ns/wl  ",
		Started: true,
	}, Policy{AsSystemRetry: true})
	require.NotNil(t, err)
	assert.Equal(t, "gang evicted by kueue (PodsReadyTimeout): Exceeded the PodsReady timeout ns/wl", err.GetMessage())

	_, err = Decide(Eviction{Started: true}, Policy{AsSystemRetry: true})
	require.NotNil(t, err)
	assert.Equal(t, "gang evicted", err.GetMessage())

	_, err = Decide(Eviction{Source: kueueSource, Reason: ReasonUnknown, Started: true}, Policy{AsSystemRetry: true})
	require.NotNil(t, err)
	assert.Equal(t, "gang evicted by kueue", err.GetMessage(), "Unknown reason is not printed")
}

func TestIsEviction(t *testing.T) {
	assert.False(t, IsEviction(nil))
	assert.True(t, IsEviction(&core.ExecutionError{Code: CodeGangEvicted, Kind: core.ExecutionError_SYSTEM}))
	userKind := &core.ExecutionError{Code: CodeGangEvicted, Kind: core.ExecutionError_USER}
	assert.False(t, IsEviction(userKind), "user-kind evictions are charged to user retries")
	assert.False(t, IsEviction(&core.ExecutionError{Code: "HostMaintenance", Kind: core.ExecutionError_SYSTEM}))
	assert.False(t, IsEviction(&core.ExecutionError{Code: CodeGangEvictionsExceeded, Kind: core.ExecutionError_SYSTEM}))
}

func TestReasonFromKueueMessage(t *testing.T) {
	// Message fixtures follow Kueue v0.19 eviction texts.
	tests := map[string]Reason{
		"Exceeded the PodsReady timeout my-ns/jobset-train-abc12": ReasonPodsReadyTimeout,
		"Preempted to accommodate a higher priority Workload":     ReasonPreempted,
		"Preempted to accommodate a workload due to reclamation":  ReasonPreempted,
		"The workload is deactivated":                             ReasonDeactivated,
		"The workload is deactivated after too many requeues":     ReasonDeactivated,
		"The ClusterQueue is stopped":                             ReasonUnknown,
		"":                                                        ReasonUnknown,
	}
	for msg, want := range tests {
		assert.Equal(t, want, ReasonFromKueueMessage(msg), msg)
	}
}

func TestWithMessage(t *testing.T) {
	assert.Nil(t, WithMessage(nil, "anything"))

	orig := &core.ExecutionError{Code: CodeGangEvicted, Kind: core.ExecutionError_SYSTEM, Message: "gang evicted by kueue"}
	same := WithMessage(orig, "   ")
	assert.Same(t, orig, same, "blank message is a no-op")

	out := WithMessage(orig, "Exceeded the PodsReady timeout ns/wl")
	require.NotSame(t, orig, out, "must not mutate the input")
	assert.Equal(t, "gang evicted by kueue", orig.GetMessage())
	assert.Equal(t, "gang evicted by kueue; gate: Exceeded the PodsReady timeout ns/wl", out.GetMessage())
	assert.Equal(t, CodeGangEvicted, out.GetCode())
	assert.Equal(t, core.ExecutionError_SYSTEM, out.GetKind())

	empty := WithMessage(&core.ExecutionError{Code: CodeGangEvicted}, "only")
	assert.Equal(t, "only", empty.GetMessage())
}

func TestBudget(t *testing.T) {
	assert.False(t, Budget{Max: 0}.Exhausted(0))
	assert.False(t, Budget{Max: 0}.Exhausted(1000), "zero max is unlimited")
	assert.False(t, Budget{Max: 3}.Exhausted(2))
	assert.True(t, Budget{Max: 3}.Exhausted(3))
	assert.True(t, Budget{Max: 3}.Exhausted(4))
}

func TestActionString(t *testing.T) {
	assert.Equal(t, "Hold", Hold.String())
	assert.Equal(t, "SystemRetry", SystemRetry.String())
	assert.Equal(t, "UserRetry", UserRetry.String())
	assert.Equal(t, "Action(7)", Action(7).String())
}
