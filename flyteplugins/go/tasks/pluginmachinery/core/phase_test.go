package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhaseInfo_WithReason(t *testing.T) {
	tests := []struct {
		name           string
		initialReason  string
		newReason      string
		expectedReason string
	}{
		{
			name:           "empty initial reason",
			initialReason:  "",
			newReason:      "new reason",
			expectedReason: "new reason",
		},
		{
			name:           "existing reason gets concatenated",
			initialReason:  "initial reason",
			newReason:      "additional reason",
			expectedReason: "initial reason, additional reason",
		},
		{
			name:           "multiple concatenations",
			initialReason:  "first reason, second reason",
			newReason:      "third reason",
			expectedReason: "first reason, second reason, third reason",
		},
		{
			name:           "empty new reason",
			initialReason:  "existing reason",
			newReason:      "",
			expectedReason: "existing reason, ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phaseInfo := PhaseInfo{
				phase:  PhaseRunning,
				reason: tt.initialReason,
			}

			phaseInfo.WithReason(tt.newReason)

			assert.Equal(t, tt.expectedReason, phaseInfo.reason)
		})
	}
}

func TestPhaseInfo_WithReason_DoesNotAffectOtherFields(t *testing.T) {
	info := &TaskInfo{}
	phaseInfo := PhaseInfo{
		phase:   PhaseRunning,
		version: 1,
		info:    info,
		reason:  "initial",
	}

	phaseInfo.WithReason("additional")

	assert.Equal(t, PhaseRunning, phaseInfo.phase)
	assert.Equal(t, uint32(1), phaseInfo.version)
	assert.Equal(t, info, phaseInfo.info)
	assert.Equal(t, "initial, additional", phaseInfo.reason)
}
