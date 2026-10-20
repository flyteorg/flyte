package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

func details(statusPhase common.ActionPhase, statusAttempts uint32, attempts ...*workflow.ActionAttempt) *workflow.ActionDetails {
	return &workflow.ActionDetails{
		Status: &workflow.ActionStatus{
			Phase:    statusPhase,
			Attempts: statusAttempts,
		},
		Attempts: attempts,
	}
}

func attempt(n uint32, phase common.ActionPhase) *workflow.ActionAttempt {
	return &workflow.ActionAttempt{Attempt: n, Phase: phase}
}

func TestActionStreamComplete_MidRetryTimeoutKeepsStreamOpen(t *testing.T) {
	// The #7910 review case: attempt 1 timed out (terminal event recorded), but the
	// action is restarting as attempt 2. Neither ordering of the two writes may
	// close the stream.

	// Event landed first: terminal TIMED_OUT for attempt 1, actions row still RUNNING.
	assert.False(t, actionStreamComplete(details(
		common.ActionPhase_ACTION_PHASE_RUNNING, 1,
		attempt(1, common.ActionPhase_ACTION_PHASE_TIMED_OUT),
	)))

	// Row moved first: action already QUEUED at attempt 2, last event still attempt 1.
	assert.False(t, actionStreamComplete(details(
		common.ActionPhase_ACTION_PHASE_QUEUED, 2,
		attempt(1, common.ActionPhase_ACTION_PHASE_TIMED_OUT),
	)))
}

func TestActionStreamComplete_TerminalTimeoutCloses(t *testing.T) {
	assert.True(t, actionStreamComplete(details(
		common.ActionPhase_ACTION_PHASE_TIMED_OUT, 2,
		attempt(1, common.ActionPhase_ACTION_PHASE_TIMED_OUT),
		attempt(2, common.ActionPhase_ACTION_PHASE_TIMED_OUT),
	)))
}

func TestActionStreamComplete_SuccessCloses(t *testing.T) {
	assert.True(t, actionStreamComplete(details(
		common.ActionPhase_ACTION_PHASE_SUCCEEDED, 1,
		attempt(1, common.ActionPhase_ACTION_PHASE_SUCCEEDED),
	)))
}

func TestActionStreamComplete_ActionTerminalButEventsLagStaysOpen(t *testing.T) {
	// The pre-existing eventual-consistency case the old predicate also guarded:
	// the actions table is terminal but action_events has not caught up yet.
	assert.False(t, actionStreamComplete(details(
		common.ActionPhase_ACTION_PHASE_FAILED, 2,
		attempt(1, common.ActionPhase_ACTION_PHASE_FAILED),
	)))
	assert.False(t, actionStreamComplete(details(
		common.ActionPhase_ACTION_PHASE_FAILED, 1,
		attempt(1, common.ActionPhase_ACTION_PHASE_RUNNING),
	)))
}

func TestActionStreamComplete_NoAttemptsStaysOpen(t *testing.T) {
	assert.False(t, actionStreamComplete(details(
		common.ActionPhase_ACTION_PHASE_SUCCEEDED, 1,
	)))
}
