package core

import (
	"context"
	"sync"
	"testing"
	"time"

	cron "github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

type recordingExecutor struct {
	mu    sync.Mutex
	calls int
}

func (r *recordingExecutor) Execute(_ context.Context, _ *models.Trigger, _ time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls++
	return nil
}

func (r *recordingExecutor) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls
}

// newCronTrigger builds a schedule trigger whose scheduling baseline (UpdatedAt)
// is `baseline`. A baseline in the past is exactly the deploy-time situation that
// used to make the job fire immediately on registration.
func newCronTrigger(t *testing.T, name, expr string, baseline time.Time) *models.Trigger {
	t.Helper()
	spec := &task.TriggerAutomationSpec{
		Type: task.TriggerAutomationSpecType_TYPE_SCHEDULE,
		Automation: &task.TriggerAutomationSpec_Schedule{
			Schedule: &task.Schedule{
				Expression: &task.Schedule_Cron{Cron: &task.Cron{Expression: expr}},
			},
		},
	}
	b, err := proto.Marshal(spec)
	require.NoError(t, err)
	return &models.Trigger{
		Project:        "p",
		Domain:         "d",
		TaskName:       "task",
		Name:           name,
		LatestRevision: 1,
		AutomationSpec: b,
		AutomationType: task.TriggerAutomationSpecType_TYPE_SCHEDULE.String(),
		UpdatedAt:      baseline,
	}
}

// TestUpdateSchedules_NoImmediateFireForPastBaseline locks in that registering a
// trigger does not schedule its first fire in the past. StartTime returns the
// deploy/last-exec baseline (in the past); the cron loop fires any entry whose
// next time is <= now immediately, so without advancing it the trigger would run
// once right after deploy. Missed runs in (lastExec, now] are handled by CatchupAll.
func TestUpdateSchedules_NoImmediateFireForPastBaseline(t *testing.T) {
	exec := &recordingExecutor{}
	s := NewGoCronScheduler(exec)

	// Hourly cron, baseline 10 minutes in the past.
	trig := newCronTrigger(t, "k1", "0 * * * *", time.Now().Add(-10*time.Minute))
	s.UpdateSchedules(context.Background(), []*models.Trigger{trig})

	require.Len(t, s.jobs, 1)
	var entryID cron.EntryID
	for _, j := range s.jobs {
		entryID = j.entryID
	}

	next := s.cron.Entry(entryID).Next
	assert.False(t, next.IsZero(), "registered job must have a next fire time")
	assert.True(t, next.After(time.Now()),
		"registered job's first fire must be strictly in the future, got %s", next)
}

// TestUpdateSchedules_RunningSchedulerDoesNotFireImmediately is the behavioral
// counterpart: start the cron loop, register a trigger with a past baseline, and
// assert the executor is not invoked promptly (the next hourly tick is far away).
func TestUpdateSchedules_RunningSchedulerDoesNotFireImmediately(t *testing.T) {
	exec := &recordingExecutor{}
	s := NewGoCronScheduler(exec)
	s.Start()
	defer s.Stop()

	trig := newCronTrigger(t, "k1", "0 * * * *", time.Now().Add(-10*time.Minute))
	s.UpdateSchedules(context.Background(), []*models.Trigger{trig})

	// Give the run loop time to process the newly added entry. With the past
	// baseline used verbatim it fired within milliseconds; the next real tick is
	// up to an hour away, so no execution should happen here.
	time.Sleep(300 * time.Millisecond)
	assert.Equal(t, 0, exec.count(), "trigger must not fire immediately on registration")
}
