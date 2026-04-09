package core

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// ParseSchedule returns a cron.Schedule for the trigger's automation spec.
// Supports both cron expressions and fixed-rate schedules.
// Returns (nil, nil) if the trigger has no schedule automation.
func ParseSchedule(t *models.Trigger) (cron.Schedule, error) {
	if t.AutomationSpec == nil {
		return nil, nil
	}
	spec := &task.TriggerAutomationSpec{}
	if err := proto.Unmarshal(t.AutomationSpec, spec); err != nil {
		return nil, err
	}
	if spec.GetType() != task.TriggerAutomationSpecType_TYPE_SCHEDULE {
		return nil, nil
	}
	sched := spec.GetSchedule()
	if sched == nil {
		return nil, nil
	}

	// Prefer the structured Cron object; fall back to legacy CronExpression string.
	if c := sched.GetCron(); c != nil && c.GetExpression() != "" {
		expr := c.GetExpression()
		parsed, err := cron.ParseStandard(expr)
		if err != nil {
			return nil, fmt.Errorf("invalid cron expression %q: %w", expr, err)
		}
		return parsed, nil
	}
	if expr := sched.GetCronExpression(); expr != "" { //nolint:staticcheck // legacy field
		parsed, err := cron.ParseStandard(expr)
		if err != nil {
			return nil, fmt.Errorf("invalid cron expression %q: %w", expr, err)
		}
		return parsed, nil
	}

	// Fixed-rate schedule.
	if rate := sched.GetRate(); rate != nil {
		d, err := fixedRateDuration(rate)
		if err != nil {
			return nil, err
		}
		return cron.ConstantDelaySchedule{Delay: d}, nil
	}

	return nil, nil
}

// fixedRateDuration converts a FixedRate proto to a time.Duration.
func fixedRateDuration(rate *task.FixedRate) (time.Duration, error) {
	d := time.Duration(rate.GetValue())
	switch rate.GetUnit() {
	case task.FixedRateUnit_FIXED_RATE_UNIT_MINUTE:
		return d * time.Minute, nil
	case task.FixedRateUnit_FIXED_RATE_UNIT_HOUR:
		return d * time.Hour, nil
	case task.FixedRateUnit_FIXED_RATE_UNIT_DAY:
		return d * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unsupported fixed rate unit %v", rate.GetUnit())
	}
}

// StartTime returns the earliest point from which to start scheduling.
// For fixed-rate schedules with a start_time, returns that start_time advanced
// past the last execution time. Otherwise falls back to TriggeredAt or DeployedAt.
func StartTime(t *models.Trigger) (time.Time, error) {
	if t.AutomationSpec == nil {
		return startTimeFallback(t), nil
	}
	spec := &task.TriggerAutomationSpec{}
	if err := proto.Unmarshal(t.AutomationSpec, spec); err != nil {
		return time.Time{}, err
	}

	rate := spec.GetSchedule().GetRate()
	if rate == nil || rate.GetStartTime() == nil {
		return startTimeFallback(t), nil
	}

	// Advance start_time past the last execution so we don't re-fire already-run slots.
	sched, err := fixedRateDuration(rate)
	if err != nil {
		return time.Time{}, err
	}
	cs := cron.ConstantDelaySchedule{Delay: sched}
	lastExec := startTimeFallback(t)
	st := rate.GetStartTime().AsTime()
	for !st.After(lastExec) {
		st = cs.Next(st)
	}
	return st, nil
}

// startTimeFallback returns the latest of TriggeredAt and UpdatedAt.
// UpdatedAt is updated when a trigger is activated/deactivated, so using it as
// the baseline ensures that re-activating a trigger does not catchup runs that
// occurred while the trigger was inactive.
func startTimeFallback(t *models.Trigger) time.Time {
	last := t.UpdatedAt
	if t.TriggeredAt.Valid && t.TriggeredAt.Time.After(last) {
		last = t.TriggeredAt.Time
	}
	return last
}

// GetCatchUpTimes returns all scheduled times in (lastExecTime, to] for a trigger.
func GetCatchUpTimes(t *models.Trigger, to time.Time) ([]time.Time, error) {
	sched, err := ParseSchedule(t)
	if err != nil || sched == nil {
		return nil, err
	}

	lastExecTime := startTimeFallback(t)
	currTime, err := StartTime(t)
	if err != nil {
		return nil, err
	}

	var times []time.Time
	for currTime.Before(to) {
		if currTime.After(lastExecTime) {
			times = append(times, currTime)
		}
		next := sched.Next(currTime)
		if next.IsZero() {
			break
		}
		currTime = next
	}
	return times, nil
}

