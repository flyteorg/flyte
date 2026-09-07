package clustered

import (
	batchv1 "k8s.io/api/batch/v1"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"

	flyteerr "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/errors"
	clusteredpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/plugins"
)

// hostMaintenanceRuleName must match the JobSet rule-name regex
// ^[A-Za-z]([A-Za-z0-9_,:]*[A-Za-z0-9_])?$ — hyphens are rejected by the webhook.
const hostMaintenanceRuleName = "restartOnHostMaintenance"

// buildFailurePolicy returns the JobSet failurePolicy from the SDK spec.
// With restart_on_host_maintenance, a disruption-evicted pod fails its Job with
// reason PodFailurePolicy (see the inner Job's podFailurePolicy in build.go) and
// the rule below restarts the whole set without charging maxRestarts. Any other
// failure takes the default RestartJobSet action, which counts toward the budget.
func buildFailurePolicy(spec *clusteredpb.ClusteredTaskSpec) (*jobsetv1alpha2.FailurePolicy, error) {
	fp := spec.GetFailurePolicy()
	if fp == nil || (fp.GetMaxRestarts() == 0 && !fp.GetRestartOnHostMaintenance()) {
		return nil, nil
	}
	maxRestarts := fp.GetMaxRestarts()
	if maxRestarts < 0 {
		return nil, flyteerr.Errorf(flyteerr.BadTaskSpecification, "failure_policy.max_restarts must be >= 0, got %d", maxRestarts)
	}
	policy := &jobsetv1alpha2.FailurePolicy{
		MaxRestarts: maxRestarts,
	}
	if fp.GetRestartOnHostMaintenance() {
		policy.Rules = []jobsetv1alpha2.FailurePolicyRule{{
			Name:                hostMaintenanceRuleName,
			Action:              jobsetv1alpha2.RestartJobSetAndIgnoreMaxRestarts,
			OnJobFailureReasons: []string{batchv1.JobReasonPodFailurePolicy},
		}}
	}
	return policy, nil
}
