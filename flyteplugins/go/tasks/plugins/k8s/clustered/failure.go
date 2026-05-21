package clustered

import (
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"

	clusteredpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/plugins"
)

// buildFailurePolicy returns the JobSet failurePolicy from the SDK spec.
// restart_on_host_maintenance (eviction free-restart via RestartJobSetAndIgnoreMaxRestarts) is deferred to PR 4.
func buildFailurePolicy(spec *clusteredpb.ClusteredTaskSpec) *jobsetv1alpha2.FailurePolicy {
	fp := spec.GetFailurePolicy()
	if fp == nil || fp.GetMaxRestarts() == 0 {
		return nil
	}
	maxRestarts := fp.GetMaxRestarts()
	return &jobsetv1alpha2.FailurePolicy{
		MaxRestarts: maxRestarts,
	}
}
