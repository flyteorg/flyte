package clustered

import (
	"context"
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/logs"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/tasklog"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

// getTaskLogs synthesizes per-rank log URLs.
//
// JobSet pod-name pattern: <jobsetName>-<replicatedJob>-<jobIdx>-<podIdx>.
// We declare a single ReplicatedJob named "workers" with Replicas=1, so jobIdx
// is always 0; podIdx == JOB_COMPLETION_INDEX == NODE_RANK in [0, totalReplicas).
//
// The primary container name is recovered from the annotation written at
// BuildResource time so we don't depend on TaskExecutionID equaling the
// container name (which it isn't, in general).
func getTaskLogs(ctx context.Context, pluginContext k8s.PluginContext, jobSet *jobsetv1alpha2.JobSet) ([]*core.TaskLog, error) {
	taskTemplate, err := pluginContext.TaskReader().Read(ctx)
	if err != nil {
		return nil, err
	}

	logPlugin, err := logs.InitializeLogPlugins(logs.GetLogConfig(), taskTemplate)
	if err != nil {
		return nil, err
	}
	if logPlugin == nil {
		return nil, nil
	}

	startTime := jobSet.CreationTimestamp.Unix()
	finishTime := time.Now().Unix()
	startTimeRFC3339 := time.Unix(startTime, 0).Format(time.RFC3339)
	finishTimeRFC3339 := time.Unix(finishTime, 0).Format(time.RFC3339)
	taskExecID := pluginContext.TaskExecutionMetadata().GetTaskExecutionID()

	var totalReplicas int32
	for _, rj := range jobSet.Spec.ReplicatedJobs {
		if rj.Name == workersReplicatedJobName {
			if rj.Template.Spec.Parallelism != nil {
				totalReplicas = *rj.Template.Spec.Parallelism
			}
			break
		}
	}

	primaryContainerName := jobSet.Annotations[primaryContainerAnnotation]

	taskLogs := make([]*core.TaskLog, 0, int(totalReplicas))
	for podIdx := int32(0); podIdx < totalReplicas; podIdx++ {
		podName := fmt.Sprintf("%s-%s-0-%d", jobSet.Name, workersReplicatedJobName, podIdx)
		output, err := logPlugin.GetTaskLogs(tasklog.Input{
			PodName:              podName,
			Namespace:            jobSet.Namespace,
			LogName:              fmt.Sprintf("rank-%d", podIdx),
			PodRFC3339StartTime:  startTimeRFC3339,
			PodRFC3339FinishTime: finishTimeRFC3339,
			PodUnixStartTime:     startTime,
			PodUnixFinishTime:    finishTime,
			TaskExecutionID:      taskExecID,
			TaskTemplate:         taskTemplate,
			ContainerName:        primaryContainerName,
		})
		if err != nil {
			return nil, err
		}
		taskLogs = append(taskLogs, output.TaskLogs...)
	}
	return taskLogs, nil
}

// getLogContext builds the structured LogContext from the JobSet's live child pods.
//
// Unlike getTaskLogs (which synthesizes templated URIs from *predicted* pod names and
// requires a pod-log template to be configured in cluster config), this uses the *real*
// pods — actual names (including the Job-assigned random suffix), namespace, primary
// container, and per-container names + process timestamps — so the console can fetch
// logs natively regardless
// of log-template config. Best-effort: returns nil on list error or when no pods are
// ready yet, leaving the templated Logs path as the fallback.
func getLogContext(ctx context.Context, pluginContext k8s.PluginContext, jobSet *jobsetv1alpha2.JobSet) *core.LogContext {
	// The plugin's K8sReader already scopes List calls to this node execution's
	// namespace and execution-id/node-id labels, so no extra filters are needed.
	podList := &v1.PodList{}
	if err := pluginContext.K8sReader().List(ctx, podList); err != nil {
		logger.Warnf(ctx, "failed to list pods for JobSet %s/%s log context: %v", jobSet.Namespace, jobSet.Name, err)
		return nil
	}

	// rank0PodName returns "<jobset>-workers-0-0"; the real pod carries an additional
	// random suffix, so match on prefix to identify the primary (rank-0) pod.
	primaryPrefix := rank0PodName(jobSet.Name)
	// The authoritative primary container name is stored on the JobSet at build time
	// (see build.go). Child pods don't carry the annotations BuildPodLogContext infers
	// from, so set it explicitly to avoid resolving to the wrong container (e.g. a sidecar).
	primaryContainerName := jobSet.Annotations[primaryContainerAnnotation]
	logCtx := &core.LogContext{Pods: make([]*core.PodLogContext, 0, len(podList.Items))}
	for i := range podList.Items {
		pod := &podList.Items[i]
		// Pending pods have no logs yet and no container statuses to build contexts from.
		if pod.Status.Phase == v1.PodPending {
			continue
		}
		if strings.HasPrefix(pod.Name, primaryPrefix) {
			logCtx.PrimaryPodName = pod.Name
		}
		podLogCtx := flytek8s.BuildPodLogContext(pod)
		if primaryContainerName != "" {
			podLogCtx.PrimaryContainerName = primaryContainerName
		}
		logCtx.Pods = append(logCtx.Pods, podLogCtx)
	}
	if len(logCtx.Pods) == 0 {
		return nil
	}
	// Guarantee PrimaryPodName references a pod in Pods: if rank-0 was pending/absent,
	// fall back to the first included pod so downstream log streaming can resolve it.
	if logCtx.PrimaryPodName == "" {
		logCtx.PrimaryPodName = logCtx.Pods[0].GetPodName()
	}
	return logCtx
}
