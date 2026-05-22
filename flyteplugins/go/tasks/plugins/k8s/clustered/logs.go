package clustered

import (
	"context"
	"fmt"
	"time"

	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/logs"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/tasklog"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

// getTaskLogs synthesizes per-rank log URLs.
// Pod name pattern: <jobsetName>-workers-0-<podIdx>
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
		if rj.Name == "workers" {
			if rj.Template.Spec.Parallelism != nil {
				totalReplicas = *rj.Template.Spec.Parallelism
			}
			break
		}
	}

	taskLogs := make([]*core.TaskLog, 0, int(totalReplicas))
	for podIdx := int32(0); podIdx < totalReplicas; podIdx++ {
		podName := fmt.Sprintf("%s-workers-0-%d", jobSet.Name, podIdx)
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
			ContainerName:        pluginContext.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(),
		})
		if err != nil {
			return nil, err
		}
		taskLogs = append(taskLogs, output.TaskLogs...)
	}
	return taskLogs, nil
}
