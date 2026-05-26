package clustered

import "regexp"

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
