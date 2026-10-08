package clustered

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/gang"
)

// Kueue names the plugin reads. Kueue is not a Go dependency: its Workload is read as
// unstructured, and these strings are Kueue's API (v1beta2, verified against v0.19.5).
const (
	// kueueQueueNameLabel is the label a user puts on the task to choose a Kueue LocalQueue.
	kueueQueueNameLabel = "kueue.x-k8s.io/queue-name"
	// kueueJobUIDLabel is the label Kueue puts on a Workload with the UID of the job it gates.
	kueueJobUIDLabel = "kueue.x-k8s.io/job-uid"
	kueueAPIVersion  = "kueue.x-k8s.io/v1beta2"

	workloadConditionEvicted       = "Evicted"
	workloadConditionPreempted     = "Preempted"
	workloadConditionQuotaReserved = "QuotaReserved"

	// Evicted reasons whose cause is the user's: their gang lost a priority contest in the
	// queue they chose, or ran past the maximum execution time they set on the job.
	evictedByPreemption           = "Preempted"
	evictedByMaximumExecutionTime = "DeactivatedDueToMaximumExecutionTimeExceeded"

	quotaReservedInadmissible = "Inadmissible"
)

// workloadForJobSet returns the Kueue Workload that gates jobSet, or nil when there is none
// (Kueue is not installed, the JobSet is not in a queue, or Kueue has not created it yet).
func workloadForJobSet(
	ctx context.Context, reader client.Reader, jobSet *jobsetv1alpha2.JobSet,
) (*unstructured.Unstructured, error) {
	if reader == nil || jobSet.UID == "" {
		return nil, nil
	}
	list := &unstructured.UnstructuredList{}
	list.SetAPIVersion(kueueAPIVersion)
	list.SetKind("WorkloadList")
	if err := reader.List(ctx, list,
		client.InNamespace(jobSet.Namespace),
		client.MatchingLabels{kueueJobUIDLabel: string(jobSet.UID)},
	); err != nil {
		return nil, err
	}
	var newest *unstructured.Unstructured
	for i := range list.Items {
		wl := &list.Items[i]
		if newest == nil || wl.GetCreationTimestamp().After(newest.GetCreationTimestamp().Time) {
			newest = wl
		}
	}
	return newest, nil
}

// workloadCondition is the part of a Workload status condition the plugin reads.
type workloadCondition struct {
	status  string
	reason  string
	message string
}

func readWorkloadCondition(wl *unstructured.Unstructured, conditionType string) (workloadCondition, bool) {
	if wl == nil {
		return workloadCondition{}, false
	}
	conditions, found, err := unstructured.NestedSlice(wl.Object, "status", "conditions")
	if err != nil || !found {
		return workloadCondition{}, false
	}
	for _, c := range conditions {
		m, ok := c.(map[string]interface{})
		if !ok || m["type"] != conditionType {
			continue
		}
		str := func(key string) string {
			s, _ := m[key].(string)
			return s
		}
		return workloadCondition{status: str("status"), reason: str("reason"), message: str("message")}, true
	}
	return workloadCondition{}, false
}

// classifyEviction describes why Kueue took back a gang that had started, from the Workload's
// Evicted condition, and whose cause it was. Only causes the user owns count against the
// task's retries; anything else, including a Workload that could not be read, does not.
func classifyEviction(wl *unstructured.Unstructured) gang.Eviction {
	eviction := gang.Eviction{Source: evictionSource}
	evicted, ok := readWorkloadCondition(wl, workloadConditionEvicted)
	if !ok || evicted.status != "True" {
		return eviction
	}
	eviction.Reason = evicted.reason
	eviction.Message = evicted.message
	switch evicted.reason {
	case evictedByPreemption:
		eviction.UserCaused = true
		if preempted, ok := readWorkloadCondition(wl, workloadConditionPreempted); ok {
			eviction.Detail = preempted.reason
		}
	case evictedByMaximumExecutionTime:
		eviction.UserCaused = true
	}
	return eviction
}

// missingLocalQueue reports the LocalQueue the JobSet names when Kueue cannot admit it because
// that queue does not exist. Only that case is a user error: a queue that exists but is stopped
// or inactive is the platform's, so the gang keeps waiting. The message is Kueue's own
// ("LocalQueue <name> doesn't exist", v0.19.5); it is compared exactly against the JobSet's
// label so a different Kueue wording is never misread as a missing queue.
func missingLocalQueue(jobSet *jobsetv1alpha2.JobSet, wl *unstructured.Unstructured) (string, bool) {
	queue := jobSet.Labels[kueueQueueNameLabel]
	if queue == "" {
		return "", false
	}
	reserved, ok := readWorkloadCondition(wl, workloadConditionQuotaReserved)
	if !ok || reserved.status != "False" || reserved.reason != quotaReservedInadmissible {
		return "", false
	}
	if reserved.message != fmt.Sprintf("LocalQueue %s doesn't exist", queue) {
		return "", false
	}
	return queue, true
}
