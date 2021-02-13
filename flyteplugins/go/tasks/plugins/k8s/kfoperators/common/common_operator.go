package common

import (
	"fmt"
	"sort"
	"time"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/tasklog"

	commonOp "github.com/kubeflow/tf-operator/pkg/apis/common/v1"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	flyteerr "github.com/lyft/flyteplugins/go/tasks/errors"
	"github.com/lyft/flyteplugins/go/tasks/logs"
	pluginsCore "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
	v1 "k8s.io/api/core/v1"
)

const (
	TensorflowTaskType = "tensorflow"
	PytorchTaskType    = "pytorch"
)

func ExtractCurrentCondition(jobConditions []commonOp.JobCondition) (commonOp.JobCondition, error) {
	if jobConditions != nil {
		sort.Slice(jobConditions, func(i, j int) bool {
			return jobConditions[i].LastTransitionTime.Time.After(jobConditions[j].LastTransitionTime.Time)
		})

		for _, jc := range jobConditions {
			if jc.Status == v1.ConditionTrue {
				return jc, nil
			}
		}
	}

	return commonOp.JobCondition{}, fmt.Errorf("found no current condition. Conditions: %+v", jobConditions)
}

func GetPhaseInfo(currentCondition commonOp.JobCondition, occurredAt time.Time,
	taskPhaseInfo pluginsCore.TaskInfo) (pluginsCore.PhaseInfo, error) {
	switch currentCondition.Type {
	case commonOp.JobCreated:
		return pluginsCore.PhaseInfoQueued(occurredAt, pluginsCore.DefaultPhaseVersion, "JobCreated"), nil
	case commonOp.JobRunning:
		return pluginsCore.PhaseInfoRunning(pluginsCore.DefaultPhaseVersion, &taskPhaseInfo), nil
	case commonOp.JobSucceeded:
		return pluginsCore.PhaseInfoSuccess(&taskPhaseInfo), nil
	case commonOp.JobFailed:
		details := fmt.Sprintf("Job failed:\n\t%v - %v", currentCondition.Reason, currentCondition.Message)
		return pluginsCore.PhaseInfoRetryableFailure(flyteerr.DownstreamSystemError, details, &taskPhaseInfo), nil
	case commonOp.JobRestarting:
		return pluginsCore.PhaseInfoRunning(pluginsCore.DefaultPhaseVersion, &taskPhaseInfo), nil
	}

	return pluginsCore.PhaseInfoUndefined, nil
}

func GetLogs(taskType string, name string, namespace string,
	workersCount int32, psReplicasCount int32, chiefReplicasCount int32) ([]*core.TaskLog, error) {
	taskLogs := make([]*core.TaskLog, 0, 10)

	logPlugin, err := logs.InitializeLogPlugins(logs.GetLogConfig())

	if err != nil {
		return nil, err
	}

	if logPlugin == nil {
		return nil, nil
	}

	if taskType == PytorchTaskType {
		masterTaskLog, masterErr := logPlugin.GetTaskLogs(
			tasklog.Input{
				PodName:   name + "-master-0",
				Namespace: namespace,
				LogName:   "master",
			},
		)
		if masterErr != nil {
			return nil, masterErr
		}
		taskLogs = append(taskLogs, masterTaskLog.TaskLogs...)
	}

	// get all workers log
	for workerIndex := int32(0); workerIndex < workersCount; workerIndex++ {
		workerLog, err := logPlugin.GetTaskLogs(tasklog.Input{
			PodName:   name + fmt.Sprintf("-worker-%d", workerIndex),
			Namespace: namespace,
		})
		if err != nil {
			return nil, err
		}
		taskLogs = append(taskLogs, workerLog.TaskLogs...)
	}
	// get all parameter servers logs
	for psReplicaIndex := int32(0); psReplicaIndex < psReplicasCount; psReplicaIndex++ {
		psReplicaLog, err := logPlugin.GetTaskLogs(tasklog.Input{
			PodName:   name + fmt.Sprintf("-psReplica-%d", psReplicaIndex),
			Namespace: namespace,
		})
		if err != nil {
			return nil, err
		}
		taskLogs = append(taskLogs, psReplicaLog.TaskLogs...)
	}
	// get chief worker log, and the max number of chief worker is 1
	if chiefReplicasCount != 0 {
		chiefReplicaLog, err := logPlugin.GetTaskLogs(tasklog.Input{
			PodName:   name + fmt.Sprintf("-chiefReplica-%d", 0),
			Namespace: namespace,
		})
		if err != nil {
			return nil, err
		}
		taskLogs = append(taskLogs, chiefReplicaLog.TaskLogs...)
	}

	return taskLogs, nil
}

func OverrideDefaultContainerName(taskCtx pluginsCore.TaskExecutionContext, podSpec *v1.PodSpec,
	defaultContainerName string) {
	// Pytorch operator forces pod to have container named 'pytorch'
	// https://github.com/kubeflow/pytorch-operator/blob/037cd1b18eb77f657f2a4bc8a8334f2a06324b57/pkg/apis/pytorch/validation/validation.go#L54-L62
	// Tensorflow operator forces pod to have container named 'tensorflow'
	// https://github.com/kubeflow/tf-operator/blob/984adc287e6fe82841e4ca282dc9a2cbb71e2d4a/pkg/apis/tensorflow/validation/validation.go#L55-L63
	// hence we have to override the name set here
	// https://github.com/lyft/flyteplugins/blob/209c52d002b4e6a39be5d175bc1046b7e631c153/go/tasks/pluginmachinery/flytek8s/container_helper.go#L116
	flyteDefaultContainerName := taskCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()
	for idx, c := range podSpec.Containers {
		if c.Name == flyteDefaultContainerName {
			podSpec.Containers[idx].Name = defaultContainerName
			return
		}
	}
}
