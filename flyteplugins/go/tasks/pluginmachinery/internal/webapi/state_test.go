package webapi

import "testing"

func TestPhase_IsTerminal(t *testing.T) {
	tests := []struct {
		p    Phase
		want bool
	}{
		{PhaseNotStarted, false},
		{PhaseAllocationTokenAcquired, false},
		{PhaseResourcesCreated, false},
		{PhaseSucceeded, true},
		{PhaseSystemFailure, true},
		{PhaseUserFailure, true},
	}
	for _, tt := range tests {
		t.Run(tt.p.String(), func(t *testing.T) {
			if got := tt.p.IsTerminal(); got != tt.want {
				t.Errorf("IsTerminal() = %v, want %v", got, tt.want)
			}
		})
	}
}
