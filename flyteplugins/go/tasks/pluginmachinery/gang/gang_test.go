package gang

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

const kueueSource = "kueue"

func TestEvictionExecutionError(t *testing.T) {
	tests := []struct {
		name     string
		policy   Policy
		wantKind core.ExecutionError_ErrorKind
	}{
		{name: "system policy is a system retry", policy: Policy{AsSystemRetry: true}, wantKind: core.ExecutionError_SYSTEM},
		{name: "user policy is a user retry", policy: Policy{AsSystemRetry: false}, wantKind: core.ExecutionError_USER},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Eviction{
				Source:  kueueSource,
				Reason:  ReasonPreempted,
				Message: "Preempted to accommodate a higher priority Workload",
			}.ExecutionError(tt.policy)
			require.NotNil(t, err)
			assert.Equal(t, CodeGangEvicted, err.GetCode())
			assert.Equal(t, tt.wantKind, err.GetKind())
			assert.Contains(t, err.GetMessage(), "gang evicted by kueue")
		})
	}
}

func TestEvictionExecutionError_MessageComposition(t *testing.T) {
	system := Policy{AsSystemRetry: true}

	err := Eviction{
		Source:  kueueSource,
		Reason:  ReasonPodsReadyTimeout,
		Message: "  Exceeded the PodsReady timeout ns/wl  ",
	}.ExecutionError(system)
	assert.Equal(t, "gang evicted by kueue (PodsReadyTimeout): Exceeded the PodsReady timeout ns/wl", err.GetMessage())

	assert.Equal(t, "gang evicted", Eviction{}.ExecutionError(system).GetMessage())

	err = Eviction{Source: kueueSource, Reason: ReasonUnknown}.ExecutionError(system)
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

func TestUsesEvictionBudget(t *testing.T) {
	sys := func(code string) *core.ExecutionError {
		return &core.ExecutionError{Kind: core.ExecutionError_SYSTEM, Code: code}
	}
	usr := func(code string) *core.ExecutionError {
		return &core.ExecutionError{Kind: core.ExecutionError_USER, Code: code}
	}
	assert.True(t, UsesEvictionBudget(sys(CodeGangEvicted)))
	assert.True(t, UsesEvictionBudget(sys(CodeGangAdmissionTimeout)))
	assert.False(t, UsesEvictionBudget(usr(CodeGangEvicted)), "a user-kind eviction is charged to the task's retries")
	assert.False(t, UsesEvictionBudget(usr(CodeGangAdmissionTimeout)))
	assert.False(t, UsesEvictionBudget(sys("ResourceDeletedExternally")))
	assert.False(t, UsesEvictionBudget(nil))
}
