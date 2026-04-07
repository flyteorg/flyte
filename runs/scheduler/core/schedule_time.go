package core

import (
	"fmt"
	"hash/fnv"
	"time"

	"github.com/robfig/cron/v3"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// CronSchedule returns the cron expression for the trigger's automation spec.
// Returns ("", false) if the trigger has no schedule automation.
func CronSchedule(t *models.Trigger) (string, bool) {
	if t.AutomationSpec == nil {
		return "", false
	}
	spec := &task.TriggerAutomationSpec{}
	if err := proto.Unmarshal(t.AutomationSpec, spec); err != nil {
		return "", false
	}
	if spec.GetType() != task.TriggerAutomationSpecType_TYPE_SCHEDULE {
		return "", false
	}
	sched := spec.GetSchedule()
	if sched == nil {
		return "", false
	}
	// Prefer the structured Cron object; fall back to legacy CronExpression string.
	if c := sched.GetCron(); c != nil && c.GetExpression() != "" {
		return c.GetExpression(), true
	}
	if expr := sched.GetCronExpression(); expr != "" { //nolint:staticcheck // legacy field
		return expr, true
	}
	return "", false
}

// parseCron parses a standard 5-field cron expression (minute-granularity).
// Supports descriptors like @hourly and timezone prefix CRON_TZ=UTC.
func parseCron(expr string) (cron.Schedule, error) {
	sched, err := cron.ParseStandard(expr)
	if err != nil {
		return nil, fmt.Errorf("invalid cron expression %q: %w", expr, err)
	}
	return sched, nil
}

// NextTime returns the next scheduled time after after for the given cron expression.
func NextTime(cronExpr string, after time.Time) (time.Time, error) {
	sched, err := parseCron(cronExpr)
	if err != nil {
		return time.Time{}, err
	}
	return sched.Next(after), nil
}

// StartTime returns the earliest point from which to start scheduling.
// If the trigger has a valid TriggeredAt, we start from there; otherwise from DeployedAt.
func StartTime(t *models.Trigger) time.Time {
	if t.TriggeredAt.Valid && !t.TriggeredAt.Time.IsZero() {
		return t.TriggeredAt.Time
	}
	return t.DeployedAt
}

// GetCatchUpTimes returns all scheduled times in (lastExecTime, to) for a trigger.
// It iterates from catchUpFrom (exclusive) and skips times <= lastExecTime.
func GetCatchUpTimes(t *models.Trigger, to time.Time) ([]time.Time, error) {
	cronExpr, ok := CronSchedule(t)
	if !ok {
		return nil, nil
	}
	sched, err := parseCron(cronExpr)
	if err != nil {
		return nil, err
	}

	lastExecTime := StartTime(t)
	currTime := lastExecTime
	var times []time.Time
	for currTime.Before(to) {
		if currTime.After(lastExecTime) {
			times = append(times, currTime)
		}
		currTime = sched.Next(currTime)
		if currTime.IsZero() {
			break
		}
	}
	return times, nil
}

// NameHash returns a deterministic run name for a scheduled trigger execution.
// Format: "trg-<10-hex-char-fnv32>", e.g. "trg-1a2b3c4d5e".
func NameHash(project, domain, taskName, triggerName string, scheduledAt time.Time) string {
	h := fnv.New64a()
	_, _ = fmt.Fprintf(h, "%s:%s:%s:%s:%d", project, domain, taskName, triggerName, scheduledAt.UnixNano())
	return fmt.Sprintf("trg-%016x", h.Sum64())[:14] // "trg-" + 10 hex chars
}
