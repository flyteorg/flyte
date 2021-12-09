package k8s

import (
	"context"
	"strconv"
	"strings"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	idlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/tasklog"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/arraystatus"
	arrayCore "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/core"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/errorcollector"
	"github.com/flyteorg/flytestdlib/bitarray"
	errors2 "github.com/flyteorg/flytestdlib/errors"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/storage"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sTypes "k8s.io/apimachinery/pkg/types"
)

type Task struct {
	State            *arrayCore.State
	NewArrayStatus   *arraystatus.ArrayStatus
	Config           *Config
	ChildIdx         int
	OriginalIndex    int
	MessageCollector *errorcollector.ErrorMessageCollector
	SubTaskIDs       []*string
}

type LaunchResult int8
type MonitorResult int8

const (
	LaunchSuccess LaunchResult = iota
	LaunchError
	LaunchWaiting
	LaunchReturnState
)

const finalizer = "flyte/array"

func addPodFinalizer(pod *corev1.Pod) *corev1.Pod {
	pod.Finalizers = append(pod.Finalizers, finalizer)
	return pod
}

func removeString(list []string, target string) []string {
	ret := make([]string, 0)
	for _, s := range list {
		if s != target {
			ret = append(ret, s)
		}
	}

	return ret
}

func clearFinalizer(pod *corev1.Pod) *corev1.Pod {
	pod.Finalizers = removeString(pod.Finalizers, finalizer)
	return pod
}

const (
	MonitorSuccess MonitorResult = iota
	MonitorError
)

func getTaskContainerIndex(pod *v1.Pod) (int, error) {
	primaryContainerName, ok := pod.Annotations[primaryContainerKey]
	// For tasks with a Container target, we only ever build one container as part of the pod
	if !ok {
		if len(pod.Spec.Containers) == 1 {
			return 0, nil
		}
		// For tasks with a K8sPod task target, they may produce multiple containers but at least one must be the designated primary.
		return -1, errors2.Errorf(ErrBuildPodTemplate, "Expected a specified primary container key when building an array job with a K8sPod spec target")

	}

	for idx, container := range pod.Spec.Containers {
		if container.Name == primaryContainerName {
			return idx, nil
		}
	}
	return -1, errors2.Errorf(ErrBuildPodTemplate, "Couldn't find any container matching the primary container key when building an array job with a K8sPod spec target")
}

func (t Task) Launch(ctx context.Context, tCtx core.TaskExecutionContext, kubeClient core.KubeClient) (LaunchResult, error) {
	podTemplate, _, err := FlyteArrayJobToK8sPodTemplate(ctx, tCtx, t.Config.NamespaceTemplate)
	if err != nil {
		return LaunchError, errors2.Wrapf(ErrBuildPodTemplate, err, "Failed to convert task template to a pod template for a task")
	}
	// Remove owner references for remote cluster execution
	if t.Config.RemoteClusterConfig.Enabled {
		podTemplate.OwnerReferences = nil
	}

	if len(podTemplate.Spec.Containers) == 0 {
		return LaunchError, errors2.Wrapf(ErrReplaceCmdTemplate, err, "No containers found in podSpec.")
	}

	containerIndex, err := getTaskContainerIndex(&podTemplate)
	if err != nil {
		return LaunchError, err
	}

	indexStr := strconv.Itoa(t.ChildIdx)
	podName := formatSubTaskName(ctx, tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), indexStr)
	allocationStatus, err := allocateResource(ctx, tCtx, t.Config, podName)
	if err != nil {
		return LaunchError, err
	}

	if allocationStatus != core.AllocationStatusGranted {
		t.NewArrayStatus.Detailed.SetItem(t.ChildIdx, bitarray.Item(core.PhaseWaitingForResources))
		t.NewArrayStatus.Summary.Inc(core.PhaseWaitingForResources)
		return LaunchWaiting, nil
	}

	pod := podTemplate.DeepCopy()
	pod.Name = podName
	pod.Spec.Containers[containerIndex].Env = append(pod.Spec.Containers[containerIndex].Env, corev1.EnvVar{
		Name: FlyteK8sArrayIndexVarName,
		// Use the OriginalIndex which represents the position of the subtask in the original user's map task before
		// compacting indexes caused by catalog-cache-check.
		Value: strconv.Itoa(t.OriginalIndex),
	})

	pod.Spec.Containers[containerIndex].Env = append(pod.Spec.Containers[containerIndex].Env, arrayJobEnvVars...)
	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return LaunchError, errors2.Wrapf(ErrGetTaskTypeVersion, err, "Unable to read task template")
	} else if taskTemplate == nil {
		return LaunchError, errors2.Wrapf(ErrGetTaskTypeVersion, err, "Missing task template")
	}

	pod = ApplyPodPolicies(ctx, t.Config, pod)
	pod = applyNodeSelectorLabels(ctx, t.Config, pod)
	pod = applyPodTolerations(ctx, t.Config, pod)
	pod = addPodFinalizer(pod)

	// Check for existing pods to prevent unnecessary Resource-Quota usage: https://github.com/kubernetes/kubernetes/issues/76787
	existingPod := &corev1.Pod{}
	err = kubeClient.GetCache().Get(ctx, client.ObjectKey{
		Namespace: pod.GetNamespace(),
		Name:      pod.GetName(),
	}, existingPod)

	if err != nil && k8serrors.IsNotFound(err) {
		// Attempt creating non-existing pod.
		err = kubeClient.GetClient().Create(ctx, pod)
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			if k8serrors.IsForbidden(err) {
				if strings.Contains(err.Error(), "exceeded quota") {
					// TODO: Quota errors are retried forever, it would be good to have support for backoff strategy.
					logger.Infof(ctx, "Failed to launch  job, resource quota exceeded. Err: %v", err)
					t.State = t.State.SetPhase(arrayCore.PhaseWaitingForResources, 0).SetReason("Not enough resources to launch job")
				} else {
					t.State = t.State.SetPhase(arrayCore.PhaseRetryableFailure, 0).SetReason("Failed to launch job.")
				}

				t.State.SetReason(err.Error())
				return LaunchReturnState, nil
			}

			return LaunchError, errors2.Wrapf(ErrSubmitJob, err, "Failed to submit job.")
		}
	} else if err != nil {
		// Another error returned.
		logger.Error(ctx, err)
		return LaunchError, errors2.Wrapf(ErrSubmitJob, err, "Failed to submit job.")
	}

	return LaunchSuccess, nil
}

func (t *Task) Monitor(ctx context.Context, tCtx core.TaskExecutionContext, kubeClient core.KubeClient, dataStore *storage.DataStore, outputPrefix, baseOutputDataSandbox storage.DataReference,
	logPlugin tasklog.Plugin) (MonitorResult, []*idlCore.TaskLog, error) {
	indexStr := strconv.Itoa(t.ChildIdx)
	podName := formatSubTaskName(ctx, tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), indexStr)
	t.SubTaskIDs = append(t.SubTaskIDs, &podName)
	var loglinks []*idlCore.TaskLog

	// Use original-index for log-name/links
	originalIdx := arrayCore.CalculateOriginalIndex(t.ChildIdx, t.State.GetIndexesToCache())
	phaseInfo, err := FetchPodStatusAndLogs(ctx, kubeClient,
		k8sTypes.NamespacedName{
			Name:      podName,
			Namespace: GetNamespaceForExecution(tCtx, t.Config.NamespaceTemplate),
		},
		originalIdx,
		tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID().RetryAttempt,
		logPlugin)
	if err != nil {
		return MonitorError, loglinks, errors2.Wrapf(ErrCheckPodStatus, err, "Failed to check pod status.")
	}

	if phaseInfo.Info() != nil {
		loglinks = phaseInfo.Info().Logs
	}

	if phaseInfo.Err() != nil {
		t.MessageCollector.Collect(t.ChildIdx, phaseInfo.Err().String())
	}

	actualPhase := phaseInfo.Phase()
	if phaseInfo.Phase().IsSuccess() {
		actualPhase, err = array.CheckTaskOutput(ctx, dataStore, outputPrefix, baseOutputDataSandbox, t.ChildIdx, originalIdx)
		if err != nil {
			return MonitorError, loglinks, err
		}
	}

	t.NewArrayStatus.Detailed.SetItem(t.ChildIdx, bitarray.Item(actualPhase))
	t.NewArrayStatus.Summary.Inc(actualPhase)

	return MonitorSuccess, loglinks, nil
}

func (t Task) Abort(ctx context.Context, tCtx core.TaskExecutionContext, kubeClient core.KubeClient) error {
	indexStr := strconv.Itoa(t.ChildIdx)
	podName := formatSubTaskName(ctx, tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), indexStr)
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       PodKind,
			APIVersion: metav1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: GetNamespaceForExecution(tCtx, t.Config.NamespaceTemplate),
		},
	}

	err := kubeClient.GetClient().Delete(ctx, pod)
	if err != nil {
		if k8serrors.IsNotFound(err) {

			return nil
		}
		return err
	}

	return nil

}

func (t Task) Finalize(ctx context.Context, tCtx core.TaskExecutionContext, kubeClient core.KubeClient) error {
	indexStr := strconv.Itoa(t.ChildIdx)
	podName := formatSubTaskName(ctx, tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), indexStr)

	pod := &v1.Pod{
		TypeMeta: metaV1.TypeMeta{
			Kind:       PodKind,
			APIVersion: v1.SchemeGroupVersion.String(),
		},
	}

	err := kubeClient.GetClient().Get(ctx, k8sTypes.NamespacedName{
		Name:      podName,
		Namespace: GetNamespaceForExecution(tCtx, t.Config.NamespaceTemplate),
	}, pod)

	if err != nil {
		if !k8serrors.IsNotFound(err) {
			logger.Errorf(ctx, "Error fetching pod [%s] in Finalize [%s]", podName, err)
			return err
		}
	} else {
		pod = clearFinalizer(pod)
		err := kubeClient.GetClient().Update(ctx, pod)
		if err != nil {
			logger.Errorf(ctx, "Error updating pod finalizer [%s] in Finalize [%s]", podName, err)
			return err
		}
	}

	// Deallocate Resource
	err = deallocateResource(ctx, tCtx, t.Config, t.ChildIdx)
	if err != nil {
		logger.Errorf(ctx, "Error releasing allocation token [%s] in Finalize [%s]", podName, err)
		return err
	}

	return nil

}

func allocateResource(ctx context.Context, tCtx core.TaskExecutionContext, config *Config, podName string) (core.AllocationStatus, error) {
	if !IsResourceConfigSet(config.ResourceConfig) {
		return core.AllocationStatusGranted, nil
	}

	resourceNamespace := core.ResourceNamespace(config.ResourceConfig.PrimaryLabel)
	resourceConstraintSpec := core.ResourceConstraintsSpec{
		ProjectScopeResourceConstraint:   nil,
		NamespaceScopeResourceConstraint: nil,
	}

	allocationStatus, err := tCtx.ResourceManager().AllocateResource(ctx, resourceNamespace, podName, resourceConstraintSpec)
	if err != nil {
		logger.Errorf(ctx, "Resource manager failed for TaskExecId [%s] token [%s]. error %s",
			tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID(), podName, err)
		return core.AllocationUndefined, err
	}

	logger.Infof(ctx, "Allocation result for [%s] is [%s]", podName, allocationStatus)
	return allocationStatus, nil
}

func deallocateResource(ctx context.Context, tCtx core.TaskExecutionContext, config *Config, childIdx int) error {
	if !IsResourceConfigSet(config.ResourceConfig) {
		return nil
	}
	indexStr := strconv.Itoa((childIdx))
	podName := formatSubTaskName(ctx, tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), indexStr)
	resourceNamespace := core.ResourceNamespace(config.ResourceConfig.PrimaryLabel)

	err := tCtx.ResourceManager().ReleaseResource(ctx, resourceNamespace, podName)
	if err != nil {
		logger.Errorf(ctx, "Error releasing token [%s]. error %s", podName, err)
		return err
	}

	return nil
}
