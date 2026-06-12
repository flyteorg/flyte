package clustered

import (
	"context"
	"regexp"
	"strings"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

// Name of the sole ReplicatedJob in the JobSet. Pod names follow the pattern
// <jobsetName>-<replicatedJob>-<jobIdx>-<podIdx>; we run a single ReplicatedJob
// with Replicas=1, so jobIdx is always 0 and podIdx == JOB_COMPLETION_INDEX == NODE_RANK.
const workersReplicatedJobName = "workers"

// Annotation set at build time so status-time code (logs, demystify) can recover
// the primary container name without re-running flytek8s.ToK8sPodSpec.
const primaryContainerAnnotation = "flyte.org/primary-container"

// (kueue): Gang admission via Kueue is not yet wired.

var (
	labelSanitizeRE = regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	labelLeadingRE  = regexp.MustCompile(`^[^a-zA-Z0-9]`)
	labelTrailingRE = regexp.MustCompile(`[^a-zA-Z0-9]$`)
)

// rank0PodName returns the unsuffixed pod name for rank 0 (jobIdx=0, podIdx=0) in the workers ReplicatedJob.
// Real Job pods carry an additional random suffix assigned by the Job controller.
func rank0PodName(jobSetName string) string {
	return jobSetName + "-" + workersReplicatedJobName + "-0-0"
}

func isRank0PodName(jobSetName, podName string) bool {
	return strings.HasPrefix(podName, rank0PodName(jobSetName))
}

func isActivePodPhase(phase v1.PodPhase) bool {
	return phase == v1.PodRunning || phase == v1.PodPending
}

// shouldPreferRank0Pod reports whether candidate should replace current as the selected rank-0 pod.
// Prefer active pods (Running/Pending) to avoid selecting stale failed pods from previous attempts.
func shouldPreferRank0Pod(candidate, current *v1.Pod) bool {
	candidateActive := isActivePodPhase(candidate.Status.Phase)
	currentActive := isActivePodPhase(current.Status.Phase)
	if candidateActive != currentActive {
		return candidateActive
	}
	if candidate.CreationTimestamp.Time.Equal(current.CreationTimestamp.Time) {
		return candidate.Name > current.Name
	}
	return candidate.CreationTimestamp.After(current.CreationTimestamp.Time)
}

// findRank0Pod locates the rank-0 pod by prefix-matching the real pod name, which includes a random suffix.
// Returns nil when not found or when listing pods fails.
func findRank0Pod(ctx context.Context, pluginContext k8s.PluginContext, jobSet *jobsetv1alpha2.JobSet) *v1.Pod {
	podList := &v1.PodList{}
	if err := pluginContext.K8sReader().List(ctx, podList, client.InNamespace(jobSet.Namespace)); err != nil {
		logger.Warnf(ctx, "failed to list pods for JobSet %s/%s rank-0 lookup: %v", jobSet.Namespace, jobSet.Name, err)
		return nil
	}

	var selected *v1.Pod
	for i := range podList.Items {
		pod := &podList.Items[i]
		if !isRank0PodName(jobSet.Name, pod.Name) {
			continue
		}
		if selected == nil || shouldPreferRank0Pod(pod, selected) {
			selected = pod
		}
	}
	return selected
}

// sanitizeLabelValue coerces an arbitrary string into a valid Kubernetes label
// value (≤63 chars, alphanumeric start/end, [a-zA-Z0-9._-] in between).
// Used for human-supplied identifiers like execution names.
func sanitizeLabelValue(value string) string {
	if value == "" {
		return "none"
	}
	s := labelSanitizeRE.ReplaceAllString(value, "-")
	if labelLeadingRE.MatchString(s) {
		s = "x" + s[1:]
	}
	if labelTrailingRE.MatchString(s) {
		s = s[:len(s)-1] + "x"
	}
	if len(s) > 63 {
		s = s[:63]
		if labelTrailingRE.MatchString(s) {
			s = s[:len(s)-1] + "x"
		}
	}
	return s
}
