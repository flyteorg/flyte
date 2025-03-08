package core

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

// Policy determines what happens when concurrency limit is reached
type Policy int

const (
	// Wait queues the execution until other executions complete
	Wait Policy = iota

	// Abort fails the execution immediately
	Abort

	// Replace terminates the oldest running execution
	Replace
)

// FromProto converts a proto policy to a core policy
func PolicyFromProto(policy admin.ConcurrencyPolicy) Policy {
	switch policy {
	case admin.ConcurrencyPolicy_WAIT:
		return Wait
	case admin.ConcurrencyPolicy_ABORT:
		return Abort
	case admin.ConcurrencyPolicy_REPLACE:
		return Replace
	default:
		return Wait // Default to Wait policy
	}
}

// ToString returns a string representation of the policy
func (p Policy) ToString() string {
	switch p {
	case Wait:
		return "WAIT"
	case Abort:
		return "ABORT"
	case Replace:
		return "REPLACE"
	default:
		return "UNKNOWN"
	}
}

// ConcurrencyLevel determines the scope of concurrency limits
type ConcurrencyLevel int

const (
	// LaunchPlan applies the limit across all versions of a launch plan
	LaunchPlan ConcurrencyLevel = iota

	// LaunchPlanVersion applies the limit only to a specific version
	LaunchPlanVersion
)

// FromProto converts a proto level to a core level
func LevelFromProto(level admin.ConcurrencyLevel) ConcurrencyLevel {
	switch level {
	case admin.ConcurrencyLevel_LAUNCH_PLAN:
		return LaunchPlan
	case admin.ConcurrencyLevel_LAUNCH_PLAN_VERSION:
		return LaunchPlanVersion
	default:
		return LaunchPlan // Default to launch plan level
	}
}

// PolicyEvaluator evaluates if an execution can proceed based on concurrency policies
type PolicyEvaluator interface {
	// CanExecute determines if an execution can proceed based on concurrency constraints
	// Returns true if execution can proceed, and an action to take if not
	CanExecute(ctx context.Context, execution models.Execution, policy *admin.SchedulerPolicy, runningCount int) (bool, string, error)
}

// DefaultPolicyEvaluator is the default implementation of PolicyEvaluator
type DefaultPolicyEvaluator struct{}

// CanExecute implements the PolicyEvaluator interface
func (d *DefaultPolicyEvaluator) CanExecute(
	ctx context.Context,
	execution models.Execution,
	policy *admin.SchedulerPolicy,
	runningCount int,
) (bool, string, error) {
	if policy == nil || runningCount < int(policy.Max) {
		return true, "", nil
	}

	corePolicy := PolicyFromProto(policy.Policy)
	switch corePolicy {
	case Wait:
		return false, "Execution queued - waiting for running executions to complete", nil
	case Abort:
		return false, "Execution aborted - maximum concurrency limit reached", nil
	case Replace:
		return true, "Execution will replace oldest running execution", nil
	default:
		return false, "Unknown concurrency policy", nil
	}
}
