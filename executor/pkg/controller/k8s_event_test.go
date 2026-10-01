package controller

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

func TestTruncateUTF8(t *testing.T) {
	longMessage := strings.Repeat("task failed ", 100)

	tests := []struct {
		name     string
		s        string
		max      int
		expected string
	}{
		{name: "empty input", s: "", max: 5, expected: ""},
		{name: "shorter than max", s: "task failed", max: 20, expected: "task failed"},
		{name: "exactly max", s: "task failed", max: 11, expected: "task failed"},
		{name: "word over max", s: "failed", max: 4, expected: "fail"},
		{name: "sentence over max", s: "the task failed after three retries", max: 15, expected: "the task failed"},
		{name: "zero max", s: "task failed", max: 0, expected: ""},
		{name: "message over note limit", s: longMessage, max: 1024, expected: longMessage[:1024]},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateUTF8(tt.s, tt.max)
			assert.Equal(t, tt.expected, got)
			assert.LessOrEqual(t, len(got), tt.max)
		})
	}
}

// testActionEvent returns an event for action a0 of run run1, attempt 1.
func testActionEvent(phase common.ActionPhase, version uint32) *workflow.ActionEvent {
	return &workflow.ActionEvent{
		Id: &common.ActionIdentifier{
			Run:  &common.RunIdentifier{Project: "flytesnacks", Domain: "development", Name: "run1"},
			Name: "a0",
		},
		Attempt: 1,
		Phase:   phase,
		Version: version,
		Cluster: "cluster1",
	}
}

func TestEventLevelOf(t *testing.T) {
	tests := []struct {
		name      string
		phase     common.ActionPhase
		version   uint32
		prevPhase common.ActionPhase
		expected  EventLevel
	}{
		{"succeeded", common.ActionPhase_ACTION_PHASE_SUCCEEDED, 0, common.ActionPhase_ACTION_PHASE_RUNNING, EventLevelTerminal},
		{"failed", common.ActionPhase_ACTION_PHASE_FAILED, 0, common.ActionPhase_ACTION_PHASE_RUNNING, EventLevelTerminal},
		{"aborted", common.ActionPhase_ACTION_PHASE_ABORTED, 0, common.ActionPhase_ACTION_PHASE_RUNNING, EventLevelTerminal},
		{"timed out", common.ActionPhase_ACTION_PHASE_TIMED_OUT, 0, common.ActionPhase_ACTION_PHASE_RUNNING, EventLevelTerminal},
		{"system retry", common.ActionPhase_ACTION_PHASE_QUEUED, systemRetryEventVersionBase, common.ActionPhase_ACTION_PHASE_RUNNING, EventLevelTerminal},
		{"just below system retry range", common.ActionPhase_ACTION_PHASE_QUEUED, systemRetryEventVersionBase - 1, common.ActionPhase_ACTION_PHASE_QUEUED, EventLevelDebug},
		{"first event", common.ActionPhase_ACTION_PHASE_QUEUED, 0, common.ActionPhase_ACTION_PHASE_UNSPECIFIED, EventLevelInfo},
		{"phase change", common.ActionPhase_ACTION_PHASE_RUNNING, 0, common.ActionPhase_ACTION_PHASE_QUEUED, EventLevelInfo},
		{"same phase, higher version", common.ActionPhase_ACTION_PHASE_RUNNING, 2, common.ActionPhase_ACTION_PHASE_RUNNING, EventLevelDebug},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := eventLevelOf(testActionEvent(tt.phase, tt.version), tt.prevPhase)
			assert.Equal(t, tt.expected, got)
		})
	}
}
