package clustered

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"
	"sigs.k8s.io/jobset/pkg/util/placement"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

// dns1035LabelMaxLength is the Kubernetes limit (RFC 1035 label) that generated
// pod, job, and service names must satisfy.
const dns1035LabelMaxLength = 63

// maxReplicasForNaming bounds the worst-case pod index we reserve room for when
// truncating the JobSet name. The name must be derivable from the generated name
// alone: the create path knows the replica count, but the lookup/abort path
// (BuildIdentityResource) has no task template, so it cannot. Reserving for the
// largest replica count we support keeps both paths producing the identical name.
// 99999 nodes is far beyond any real distributed-training job.
const maxReplicasForNaming = 100000

// jobSetNameSuffixLen is the number of characters JobSet appends to the JobSet name
// when deriving its longest child pod name. placement.GenPodName("", ...) yields
// exactly the "-<replicatedJob>-<jobIdx>-<podIdx>" portion JobSet appends to the
// JobSet name; "-abcde" mirrors the 5-char random suffix the Job controller adds.
// This matches what the JobSet webhook validates against.
func jobSetNameSuffixLen() int {
	maxPodIdx := strconv.Itoa(maxReplicasForNaming - 1)
	return len(placement.GenPodName("", workersReplicatedJobName, "0", maxPodIdx)) + len("-abcde")
}

// generatedNameMaxLength bounds the task's generated name via the plugin's
// PluginProperties. The plugin manager stamps the object name with
// GetGeneratedNameWith(0, GeneratedNameMaxLength) — DNS-1035-sanitized and bounded —
// on both the create and lookup paths. JobSet's admission webhook computes the
// longest child pod name as "<jobSetName>-<replicatedJob>-<jobIdx>-<podIdx>-<5-char
// random suffix>" and rejects the entire JobSet if it would exceed 63 characters;
// composed/nested tasks produce long generated names, so without this bound the
// webhook rejects the pods, the plugin retries forever, and the execution is stuck
// in RUNNING. jobIdx is always 0 (single ReplicatedJob with Replicas=1); podIdx is
// reserved for the worst case (see maxReplicasForNaming) so the bound is independent
// of the replica count.
var generatedNameMaxLength = dns1035LabelMaxLength - jobSetNameSuffixLen()

// Name of the sole ReplicatedJob in the JobSet. Pod names follow the pattern
// <jobsetName>-<replicatedJob>-<jobIdx>-<podIdx>; we run a single ReplicatedJob
// with Replicas=1, so jobIdx is always 0 and podIdx == JOB_COMPLETION_INDEX == NODE_RANK.
const workersReplicatedJobName = "workers"

// Annotation set at build time so status-time code (logs, demystify) can recover
// the primary container name without re-running flytek8s.ToK8sPodSpec.
const primaryContainerAnnotation = "flyte.org/primary-container"

// Gang admission: a JobSet held suspended by an external admission gate (Kueue's
// JobSet integration flips spec.suspend) is reported as WaitingForResources until the
// gate releases it, and as a GangEvicted system retry if the gate revokes a gang that
// had already fully started. See phase.go. Build-time wiring (queue label and
// spec.suspend on creation) is gated by plugin config.

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

// listWorkerPods lists this JobSet's worker pods. The JobSet controller labels every pod
// it creates with the JobSet and ReplicatedJob names, so the selector matches exactly
// this JobSet's workers. Returns nil when listing fails.
func listWorkerPods(ctx context.Context, pluginContext k8s.PluginContext, jobSet *jobsetv1alpha2.JobSet) []v1.Pod {
	podList := &v1.PodList{}
	if err := pluginContext.K8sReader().List(ctx, podList,
		client.InNamespace(jobSet.Namespace),
		client.MatchingLabels{
			jobsetv1alpha2.JobSetNameKey:        jobSet.Name,
			jobsetv1alpha2.ReplicatedJobNameKey: workersReplicatedJobName,
		},
	); err != nil {
		logger.Warnf(ctx, "failed to list worker pods for JobSet %s/%s: %v", jobSet.Namespace, jobSet.Name, err)
		return nil
	}
	return podList.Items
}

// selectRank0Pod picks the rank-0 pod among the worker pods by prefix-matching the real
// pod name, which carries a random suffix assigned by the Job controller. Returns nil
// when there is none.
func selectRank0Pod(jobSet *jobsetv1alpha2.JobSet, pods []v1.Pod) *v1.Pod {
	var selected *v1.Pod
	for i := range pods {
		pod := &pods[i]
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
