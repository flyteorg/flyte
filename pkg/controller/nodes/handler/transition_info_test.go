package handler

import (
	"testing"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

func TestPhaseInfoQueued(t *testing.T) {
	p := PhaseInfoQueued("queued")
	assert.Equal(t, EPhaseQueued, p.p)
}

func TestEPhase_String(t *testing.T) {
	tests := []struct {
		name string
		p    EPhase
	}{
		{"queued", EPhaseQueued},
		{"not-ready", EPhaseNotReady},
		{"timedout", EPhaseTimedout},
		{"undefined", EPhaseUndefined},
		{"success", EPhaseSuccess},
		{"skip", EPhaseSkip},
		{"failed", EPhaseFailed},
		{"running", EPhaseRunning},
		{"retryable-fail", EPhaseRetryableFailure},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.String(); got != tt.name {
				t.Errorf("String() = %v, want %v", got, tt.name)
			}
		})
	}
}

func TestEPhase_IsTerminal(t *testing.T) {
	tests := []struct {
		name string
		p    EPhase
		want bool
	}{
		{"success", EPhaseSuccess, true},
		{"failure", EPhaseFailed, true},
		{"timeout", EPhaseTimedout, true},
		{"skip", EPhaseSkip, true},
		{"any", EPhaseQueued, false},
		{"retryable", EPhaseRetryableFailure, false},
		{"run", EPhaseRunning, false},
		{"nr", EPhaseNotReady, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.IsTerminal(); got != tt.want {
				t.Errorf("IsTerminal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPhaseInfo(t *testing.T) {
	t.Run("undefined", func(t *testing.T) {
		assert.Equal(t, EPhaseUndefined, PhaseInfoUndefined.GetPhase())
	})

	t.Run("success", func(t *testing.T) {
		i := &ExecutionInfo{}
		p := PhaseInfoSuccess(i)
		assert.Equal(t, EPhaseSuccess, p.GetPhase())
		assert.Equal(t, i, p.GetInfo())
		assert.Nil(t, p.GetErr())
		assert.NotNil(t, p.GetOccurredAt())
	})

	t.Run("not-ready", func(t *testing.T) {
		p := PhaseInfoNotReady("reason")
		assert.Equal(t, EPhaseNotReady, p.GetPhase())
		assert.Nil(t, p.GetErr())
		assert.NotNil(t, p.GetOccurredAt())
		assert.Equal(t, "reason", p.GetReason())
	})

	t.Run("queued", func(t *testing.T) {
		p := PhaseInfoQueued("reason")
		assert.Equal(t, EPhaseQueued, p.GetPhase())
		assert.Nil(t, p.GetErr())
		assert.NotNil(t, p.GetOccurredAt())
		assert.Equal(t, "reason", p.GetReason())
	})

	t.Run("running", func(t *testing.T) {
		i := &ExecutionInfo{}
		p := PhaseInfoRunning(i)
		assert.Equal(t, EPhaseRunning, p.GetPhase())
		assert.Equal(t, i, p.GetInfo())
		assert.Nil(t, p.GetErr())
		assert.NotNil(t, p.GetOccurredAt())
	})

	t.Run("skip", func(t *testing.T) {
		i := &ExecutionInfo{}
		p := PhaseInfoSkip(i, "reason")
		assert.Equal(t, EPhaseSkip, p.GetPhase())
		assert.Equal(t, i, p.GetInfo())
		assert.Nil(t, p.GetErr())
		assert.NotNil(t, p.GetOccurredAt())
		assert.Equal(t, "reason", p.GetReason())
	})

	t.Run("timeout", func(t *testing.T) {
		i := &ExecutionInfo{}
		p := PhaseInfoTimedOut(i, "reason")
		assert.Equal(t, EPhaseTimedout, p.GetPhase())
		assert.Equal(t, i, p.GetInfo())
		assert.Nil(t, p.GetErr())
		assert.NotNil(t, p.GetOccurredAt())
		assert.Equal(t, "reason", p.GetReason())
	})

	t.Run("failure", func(t *testing.T) {
		i := &ExecutionInfo{}
		p := PhaseInfoFailure("code", "reason", i)
		assert.Equal(t, EPhaseFailed, p.GetPhase())
		assert.Equal(t, i, p.GetInfo())
		if assert.NotNil(t, p.GetErr()) {
			assert.Equal(t, "code", p.GetErr().Code)
			assert.Equal(t, "reason", p.GetErr().Message)
		}
		assert.NotNil(t, p.GetOccurredAt())
	})

	t.Run("failure-err", func(t *testing.T) {
		i := &ExecutionInfo{}
		e := &core.ExecutionError{}
		p := PhaseInfoFailureErr(e, i)
		assert.Equal(t, EPhaseFailed, p.GetPhase())
		assert.Equal(t, i, p.GetInfo())
		assert.Equal(t, e, p.GetErr())
		assert.NotNil(t, p.GetOccurredAt())
	})

	t.Run("failure-err", func(t *testing.T) {
		i := &ExecutionInfo{}
		p := PhaseInfoFailureErr(nil, i)
		assert.Equal(t, EPhaseFailed, p.GetPhase())
		assert.Equal(t, i, p.GetInfo())
		assert.NotNil(t, p.GetErr())
		assert.NotNil(t, p.GetOccurredAt())
	})

	t.Run("retryable-fail", func(t *testing.T) {
		i := &ExecutionInfo{}
		p := PhaseInfoRetryableFailure("code", "reason", i)
		assert.Equal(t, EPhaseRetryableFailure, p.GetPhase())
		assert.Equal(t, i, p.GetInfo())
		if assert.NotNil(t, p.GetErr()) {
			assert.Equal(t, "code", p.GetErr().Code)
			assert.Equal(t, "reason", p.GetErr().Message)
		}
		assert.NotNil(t, p.GetOccurredAt())
	})

	t.Run("retryable-fail-err", func(t *testing.T) {
		i := &ExecutionInfo{}
		e := &core.ExecutionError{}
		p := PhaseInfoRetryableFailureErr(e, i)
		assert.Equal(t, EPhaseRetryableFailure, p.GetPhase())
		assert.Equal(t, i, p.GetInfo())
		assert.Equal(t, e, p.GetErr())
		assert.NotNil(t, p.GetOccurredAt())
	})
}
