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

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/encoding"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	stdlibutils "github.com/flyteorg/flyte/v2/flytestdlib/utils"
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

// generatedNameMaxLength bounds the task's generated name itself, via the plugin's
// PluginProperties. Plugin managers may stamp GetGeneratedName() directly onto the
// object they create — overwriting the name BuildResource chose — so truncating
// inside buildJobSetName alone is not enough: the bound has to hold at the source.
// Exposing it as GeneratedNameMaxLength makes every execution context that honors
// the property produce a generated name whose derived child pod names fit the
// 63-char limit, and buildJobSetName then passes it through unchanged so the
// create, lookup, and any manager-stamped paths all agree on the same name.
var generatedNameMaxLength = dns1035LabelMaxLength - jobSetNameSuffixLen()

// buildJobSetName derives the JobSet name from the task's generated name while
// guaranteeing that the longest child pod name JobSet generates stays within the
// 63-character DNS-1035 label limit.
//
// JobSet's admission webhook computes the longest pod name as
// "<jobSetName>-<replicatedJob>-<jobIdx>-<podIdx>-<5-char random suffix>" and
// rejects the entire JobSet if it would exceed 63 characters. Composed/nested
// tasks produce long generated names, so without truncating here the webhook
// rejects the pods, the plugin retries forever, and the execution is stuck in
// RUNNING. Generated names are normally already bounded at the source (see
// generatedNameMaxLength), making this a defense-in-depth no-op; it still truncates
// with a hash for contexts that don't honor GeneratedNameMaxLength. jobIdx is
// always 0 (single ReplicatedJob with Replicas=1); podIdx is reserved for the worst
// case (see maxReplicasForNaming) so the result is independent of the replica count
// and matches on both the create and lookup paths.
func buildJobSetName(generatedName string) string {
	name := toDNS1035Label(generatedName)

	maxLen := dns1035LabelMaxLength - jobSetNameSuffixLen()
	if maxLen < 1 {
		maxLen = 1
	}
	if len(name) <= maxLen {
		return name
	}

	if fixedLengthID, err := encoding.FixedLengthUniqueID(name, stdlibutils.MaxUniqueIDLength); err == nil && maxLen > len(fixedLengthID)+1 {
		prefix := strings.TrimRight(name[:maxLen-len(fixedLengthID)-1], "-")
		return prefix + "-" + fixedLengthID
	}
	return strings.TrimRight(name[:maxLen], "-")
}

// toDNS1035Label coerces an arbitrary string into a valid RFC 1035 DNS label:
// starts with a letter, ends with an alphanumeric, and contains only [-a-z0-9]
// (no dots). The JobSet name becomes a component of the child Job/Pod names, which
// JobSet's admission webhook validates as DNS-1035 labels. ConvertToDNS1123Subdomain-
// CompatibleString only targets the looser subdomain rules and can legally retain a
// '.' or a leading digit, either of which would make the derived label invalid and be
// rejected by the webhook — so we tighten it here.
func toDNS1035Label(name string) string {
	name = utils.ConvertToDNS1123SubdomainCompatibleString(name)
	name = strings.ReplaceAll(name, ".", "-")
	name = strings.Trim(name, "-")
	if name == "" || name[0] < 'a' || name[0] > 'z' {
		name = "x" + name
	}
	return name
}

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

// findRank0Pod locates the rank-0 pod among this execution's child pods.
//
// The plugin's K8sReader already scopes List calls to this node execution's namespace and
// execution-id/node-id labels (propagated onto the pod template in build.go), so the returned
// pods belong to this JobSet. We keep an explicit namespace filter as defense-in-depth and then
// prefix-match the real pod name, which carries a random suffix assigned by the Job controller.
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
