package handler

import (
	"testing"

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
		{"undefined", EPhaseUndefined},
		{"success", EPhaseSuccess},
		{"skip", EPhaseSkip},
		{"failed", EPhaseFailed},
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
