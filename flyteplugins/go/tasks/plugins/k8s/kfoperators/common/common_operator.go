package common

import (
	"fmt"
	"sort"
	"time"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/tasklog"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	kfplugins "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins/kubeflow"
	flyteerr "github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyteplugins/go/tasks/logs"
	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	commonOp "github.com/kubeflow/common/pkg/apis/common/v1"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	TensorflowTaskType = "tensorflow"
	MPITaskType        = "mpi"
	PytorchTaskType    = "pytorch"
)

type ReplicaEntry struct {
	PodSpec       *v1.PodSpec
	ReplicaNum    int32
	RestartPolicy commonOp.RestartPolicy
}

// ExtractCurrentCondition will return the first job condition for tensorflow/pytorch
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
		return commonOp.JobCondition{}, fmt.Errorf("found no current condition. Conditions: %+v", jobConditions)
	}
	return commonOp.JobCondition{}, nil
}

// GetPhaseInfo will return the phase of kubeflow job
func GetPhaseInfo(currentCondition commonOp.JobCondition, occurredAt time.Time,
	taskPhaseInfo pluginsCore.TaskInfo) (pluginsCore.PhaseInfo, error) {
	if len(currentCondition.Type) == 0 {
		return pluginsCore.PhaseInfoQueued(occurredAt, pluginsCore.DefaultPhaseVersion, "JobCreated"), nil
	}
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

// GetMPIPhaseInfo will return the phase of MPI job
func GetMPIPhaseInfo(currentCondition commonOp.JobCondition, occurredAt time.Time,
	taskPhaseInfo pluginsCore.TaskInfo) (pluginsCore.PhaseInfo, error) {
	switch currentCondition.Type {
	case commonOp.JobCreated:
		return pluginsCore.PhaseInfoQueued(occurredAt, pluginsCore.DefaultPhaseVersion, "New job name submitted to MPI operator"), nil
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

// GetLogs will return the logs for kubeflow job
func GetLogs(pluginContext k8s.PluginContext, taskType string, objectMeta meta_v1.ObjectMeta, hasMaster bool,
	workersCount int32, psReplicasCount int32, chiefReplicasCount int32) ([]*core.TaskLog, error) {
	name := objectMeta.Name
	namespace := objectMeta.Namespace

	taskLogs := make([]*core.TaskLog, 0, 10)
	taskExecID := pluginContext.TaskExecutionMetadata().GetTaskExecutionID().GetID()

	logPlugin, err := logs.InitializeLogPlugins(logs.GetLogConfig())

	if err != nil {
		return nil, err
	}

	if logPlugin == nil {
		return nil, nil
	}

	// We use the creation timestamp of the Kubeflow Job as a proxy for the start time of the pods
	startTime := objectMeta.CreationTimestamp.Time.Unix()
	// Don't have a good mechanism for this yet, but approximating with time.Now for now
	finishTime := time.Now().Unix()
	RFC3999StartTime := time.Unix(startTime, 0).Format(time.RFC3339)
	RFC3999FinishTime := time.Unix(finishTime, 0).Format(time.RFC3339)

	if taskType == PytorchTaskType && hasMaster {
		masterTaskLog, masterErr := logPlugin.GetTaskLogs(
			tasklog.Input{
				PodName:                 name + "-master-0",
				Namespace:               namespace,
				LogName:                 "master",
				PodRFC3339StartTime:     RFC3999StartTime,
				PodRFC3339FinishTime:    RFC3999FinishTime,
				PodUnixStartTime:        startTime,
				PodUnixFinishTime:       finishTime,
				TaskExecutionIdentifier: &taskExecID,
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
			PodName:                 name + fmt.Sprintf("-worker-%d", workerIndex),
			Namespace:               namespace,
			PodRFC3339StartTime:     RFC3999StartTime,
			PodRFC3339FinishTime:    RFC3999FinishTime,
			PodUnixStartTime:        startTime,
			PodUnixFinishTime:       finishTime,
			TaskExecutionIdentifier: &taskExecID,
		})
		if err != nil {
			return nil, err
		}
		taskLogs = append(taskLogs, workerLog.TaskLogs...)
	}

	if taskType == MPITaskType || taskType == PytorchTaskType {
		return taskLogs, nil
	}

	// get all parameter servers logs
	for psReplicaIndex := int32(0); psReplicaIndex < psReplicasCount; psReplicaIndex++ {
		psReplicaLog, err := logPlugin.GetTaskLogs(tasklog.Input{
			PodName:                 name + fmt.Sprintf("-psReplica-%d", psReplicaIndex),
			Namespace:               namespace,
			TaskExecutionIdentifier: &taskExecID,
		})
		if err != nil {
			return nil, err
		}
		taskLogs = append(taskLogs, psReplicaLog.TaskLogs...)
	}
	// get chief worker log, and the max number of chief worker is 1
	if chiefReplicasCount != 0 {
		chiefReplicaLog, err := logPlugin.GetTaskLogs(tasklog.Input{
			PodName:                 name + fmt.Sprintf("-chiefReplica-%d", 0),
			Namespace:               namespace,
			TaskExecutionIdentifier: &taskExecID,
		})
		if err != nil {
			return nil, err
		}
		taskLogs = append(taskLogs, chiefReplicaLog.TaskLogs...)
	}

	return taskLogs, nil
}

func OverridePrimaryContainerName(podSpec *v1.PodSpec, primaryContainerName string, defaultContainerName string) {
	// Pytorch operator forces pod to have container named 'pytorch'
	// https://github.com/kubeflow/pytorch-operator/blob/037cd1b18eb77f657f2a4bc8a8334f2a06324b57/pkg/apis/pytorch/validation/validation.go#L54-L62
	// Tensorflow operator forces pod to have container named 'tensorflow'
	// https://github.com/kubeflow/tf-operator/blob/984adc287e6fe82841e4ca282dc9a2cbb71e2d4a/pkg/apis/tensorflow/validation/validation.go#L55-L63
	// hence we have to override the name set here
	// https://github.com/flyteorg/flyteplugins/blob/209c52d002b4e6a39be5d175bc1046b7e631c153/go/tasks/pluginmachinery/flytek8s/container_helper.go#L116
	for idx, c := range podSpec.Containers {
		if c.Name == primaryContainerName {
			podSpec.Containers[idx].Name = defaultContainerName
			return
		}
	}
}

// ParseRunPolicy converts a kubeflow plugin RunPolicy object to a k8s RunPolicy object.
func ParseRunPolicy(flyteRunPolicy kfplugins.RunPolicy) commonOp.RunPolicy {
	runPolicy := commonOp.RunPolicy{}
	if flyteRunPolicy.GetBackoffLimit() != 0 {
		var backoffLimit = flyteRunPolicy.GetBackoffLimit()
		runPolicy.BackoffLimit = &backoffLimit
	}
	var cleanPodPolicy = ParseCleanPodPolicy(flyteRunPolicy.GetCleanPodPolicy())
	runPolicy.CleanPodPolicy = &cleanPodPolicy
	if flyteRunPolicy.GetActiveDeadlineSeconds() != 0 {
		var ddlSeconds = int64(flyteRunPolicy.GetActiveDeadlineSeconds())
		runPolicy.ActiveDeadlineSeconds = &ddlSeconds
	}
	if flyteRunPolicy.GetTtlSecondsAfterFinished() != 0 {
		var ttl = flyteRunPolicy.GetTtlSecondsAfterFinished()
		runPolicy.TTLSecondsAfterFinished = &ttl
	}

	return runPolicy
}

// Get k8s clean pod policy from flyte kubeflow plugins clean pod policy.
func ParseCleanPodPolicy(flyteCleanPodPolicy kfplugins.CleanPodPolicy) commonOp.CleanPodPolicy {
	cleanPodPolicyMap := map[kfplugins.CleanPodPolicy]commonOp.CleanPodPolicy{
		kfplugins.CleanPodPolicy_CLEANPOD_POLICY_NONE:    commonOp.CleanPodPolicyNone,
		kfplugins.CleanPodPolicy_CLEANPOD_POLICY_ALL:     commonOp.CleanPodPolicyAll,
		kfplugins.CleanPodPolicy_CLEANPOD_POLICY_RUNNING: commonOp.CleanPodPolicyRunning,
	}
	return cleanPodPolicyMap[flyteCleanPodPolicy]
}

// Get k8s restart policy from flyte kubeflow plugins restart policy.
func ParseRestartPolicy(flyteRestartPolicy kfplugins.RestartPolicy) commonOp.RestartPolicy {
	restartPolicyMap := map[kfplugins.RestartPolicy]commonOp.RestartPolicy{
		kfplugins.RestartPolicy_RESTART_POLICY_NEVER:      commonOp.RestartPolicyNever,
		kfplugins.RestartPolicy_RESTART_POLICY_ON_FAILURE: commonOp.RestartPolicyOnFailure,
		kfplugins.RestartPolicy_RESTART_POLICY_ALWAYS:     commonOp.RestartPolicyAlways,
	}
	return restartPolicyMap[flyteRestartPolicy]
}

// OverrideContainerSpec overrides the specified container's properties in the given podSpec. The function
// updates the image, resources and command arguments of the container that matches the given containerName.
func OverrideContainerSpec(podSpec *v1.PodSpec, containerName string, image string, resources *core.Resources, args []string) error {
	for idx, c := range podSpec.Containers {
		if c.Name == containerName {
			if image != "" {
				podSpec.Containers[idx].Image = image
			}
			if resources != nil {
				// if resources requests and limits both not set, we will not override the resources
				if len(resources.Requests) >= 1 || len(resources.Limits) >= 1 {
					resources, err := flytek8s.ToK8sResourceRequirements(resources)
					if err != nil {
						return flyteerr.Errorf(flyteerr.BadTaskSpecification, "invalid TaskSpecification on Resources [%v], Err: [%v]", resources, err.Error())
					}
					podSpec.Containers[idx].Resources = *resources
				}
			}
			if len(args) != 0 {
				podSpec.Containers[idx].Args = args
			}
		}
	}
	return nil
}
