package k8s

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/flyteorg/flyteplugins/go/tasks/errors"
	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/tasklog"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"
	podPlugin "github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/pod"

	stdErrors "github.com/flyteorg/flytestdlib/errors"
	"github.com/flyteorg/flytestdlib/logger"

	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ErrBuildPodTemplate       stdErrors.ErrorCode = "POD_TEMPLATE_FAILED"
	ErrReplaceCmdTemplate     stdErrors.ErrorCode = "CMD_TEMPLATE_FAILED"
	FlyteK8sArrayIndexVarName string              = "FLYTE_K8S_ARRAY_INDEX"
	finalizer                 string              = "flyte/array"
	JobIndexVarName           string              = "BATCH_JOB_ARRAY_INDEX_VAR_NAME"
)

var (
	arrayJobEnvVars = []v1.EnvVar{
		{
			Name:  JobIndexVarName,
			Value: FlyteK8sArrayIndexVarName,
		},
	}
	namespaceRegex = regexp.MustCompile("(?i){{.namespace}}(?i)")
)

// addMetadata sets k8s pod metadata that is either specifically required by the k8s array plugin
// or defined in the plugin configuration.
func addMetadata(stCtx SubTaskExecutionContext, cfg *Config, k8sPluginCfg *config.K8sPluginConfig, pod *v1.Pod) {
	taskExecutionMetadata := stCtx.TaskExecutionMetadata()

	// Default to parent namespace
	namespace := taskExecutionMetadata.GetNamespace()
	if cfg.NamespaceTemplate != "" {
		if namespaceRegex.MatchString(cfg.NamespaceTemplate) {
			namespace = namespaceRegex.ReplaceAllString(cfg.NamespaceTemplate, namespace)
		} else {
			namespace = cfg.NamespaceTemplate
		}
	}

	pod.SetNamespace(namespace)
	pod.SetAnnotations(utils.UnionMaps(k8sPluginCfg.DefaultAnnotations, pod.GetAnnotations(), utils.CopyMap(taskExecutionMetadata.GetAnnotations())))
	pod.SetLabels(utils.UnionMaps(k8sPluginCfg.DefaultLabels, pod.GetLabels(), utils.CopyMap(taskExecutionMetadata.GetLabels())))
	pod.SetName(taskExecutionMetadata.GetTaskExecutionID().GetGeneratedName())

	if !cfg.RemoteClusterConfig.Enabled {
		pod.OwnerReferences = []metav1.OwnerReference{taskExecutionMetadata.GetOwnerReference()}
	}

	if k8sPluginCfg.InjectFinalizer {
		f := append(pod.GetFinalizers(), finalizer)
		pod.SetFinalizers(f)
	}

	if len(cfg.DefaultScheduler) > 0 {
		pod.Spec.SchedulerName = cfg.DefaultScheduler
	}

	// The legacy map task implemented these as overrides so they were left as such. May want to
	// revist whether they would serve better as appends.
	if len(cfg.NodeSelector) != 0 {
		pod.Spec.NodeSelector = cfg.NodeSelector
	}
	if len(cfg.Tolerations) != 0 {
		pod.Spec.Tolerations = cfg.Tolerations
	}
}

// abortSubtask attempts to interrupt the k8s pod defined by the SubTaskExecutionContext and Config
func abortSubtask(ctx context.Context, stCtx SubTaskExecutionContext, cfg *Config, kubeClient pluginsCore.KubeClient) error {
	logger.Infof(ctx, "KillTask invoked. We will attempt to delete object [%v].",
		stCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName())

	var plugin k8s.Plugin = podPlugin.DefaultPodPlugin
	o, err := plugin.BuildIdentityResource(ctx, stCtx.TaskExecutionMetadata())
	if err != nil {
		// This will recurrent, so we will skip further finalize
		logger.Errorf(ctx, "Failed to build the Resource with name: %v. Error: %v, when finalizing.", stCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), err)
		return nil
	}

	addMetadata(stCtx, cfg, config.GetK8sPluginConfig(), o.(*v1.Pod))

	deleteResource := true
	abortOverride, hasAbortOverride := plugin.(k8s.PluginAbortOverride)

	resourceToFinalize := o
	var behavior k8s.AbortBehavior

	if hasAbortOverride {
		behavior, err = abortOverride.OnAbort(ctx, stCtx, o)
		deleteResource = err == nil && behavior.DeleteResource
		if err == nil && behavior.Resource != nil {
			resourceToFinalize = behavior.Resource
		}
	}

	if err != nil {
	} else if deleteResource {
		err = kubeClient.GetClient().Delete(ctx, resourceToFinalize)
	} else {
		if behavior.Patch != nil && behavior.Update == nil {
			err = kubeClient.GetClient().Patch(ctx, resourceToFinalize, behavior.Patch.Patch, behavior.Patch.Options...)
		} else if behavior.Patch == nil && behavior.Update != nil {
			err = kubeClient.GetClient().Update(ctx, resourceToFinalize, behavior.Update.Options...)
		} else {
			err = errors.Errorf(errors.RuntimeFailure, "AbortBehavior for resource %v must specify either a Patch and an Update operation if Delete is set to false. Only one can be supplied.", resourceToFinalize.GetName())
		}
		if behavior.DeleteOnErr && err != nil {
			logger.Warningf(ctx, "Failed to apply AbortBehavior for resource %v with error %v. Will attempt to delete resource.", resourceToFinalize.GetName(), err)
			err = kubeClient.GetClient().Delete(ctx, resourceToFinalize)
		}
	}

	if err != nil && !isK8sObjectNotExists(err) {
		logger.Warningf(ctx, "Failed to clear finalizers for Resource with name: %v/%v. Error: %v",
			resourceToFinalize.GetNamespace(), resourceToFinalize.GetName(), err)
		return err
	}

	return nil
}

// clearFinalizers removes finalizers (if they exist) from the k8s resource
func clearFinalizers(ctx context.Context, o client.Object, kubeClient pluginsCore.KubeClient) error {
	if len(o.GetFinalizers()) > 0 {
		o.SetFinalizers([]string{})
		err := kubeClient.GetClient().Update(ctx, o)
		if err != nil && !isK8sObjectNotExists(err) {
			logger.Warningf(ctx, "Failed to clear finalizers for Resource with name: %v/%v. Error: %v", o.GetNamespace(), o.GetName(), err)
			return err
		}
	} else {
		logger.Debugf(ctx, "Finalizers are already empty for Resource with name: %v/%v", o.GetNamespace(), o.GetName())
	}
	return nil
}

// launchSubtask creates a k8s pod defined by the SubTaskExecutionContext and Config.
func launchSubtask(ctx context.Context, stCtx SubTaskExecutionContext, cfg *Config, kubeClient pluginsCore.KubeClient) (pluginsCore.PhaseInfo, error) {
	o, err := podPlugin.DefaultPodPlugin.BuildResource(ctx, stCtx)
	pod := o.(*v1.Pod)
	if err != nil {
		return pluginsCore.PhaseInfoUndefined, err
	}

	addMetadata(stCtx, cfg, config.GetK8sPluginConfig(), pod)

	// inject maptask specific container environment variables
	if len(pod.Spec.Containers) == 0 {
		return pluginsCore.PhaseInfoUndefined, stdErrors.Wrapf(ErrReplaceCmdTemplate, err, "No containers found in podSpec.")
	}

	containerIndex, err := getTaskContainerIndex(pod)
	if err != nil {
		return pluginsCore.PhaseInfoUndefined, err
	}

	pod.Spec.Containers[containerIndex].Env = append(pod.Spec.Containers[containerIndex].Env, v1.EnvVar{
		Name: FlyteK8sArrayIndexVarName,
		// Use the OriginalIndex which represents the position of the subtask in the original user's map task before
		// compacting indexes caused by catalog-cache-check.
		Value: strconv.Itoa(stCtx.originalIndex),
	})

	pod.Spec.Containers[containerIndex].Env = append(pod.Spec.Containers[containerIndex].Env, arrayJobEnvVars...)

	logger.Infof(ctx, "Creating Object: Type:[%v], Object:[%v/%v]", pod.GetObjectKind().GroupVersionKind(), pod.GetNamespace(), pod.GetName())
	err = kubeClient.GetClient().Create(ctx, pod)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		if k8serrors.IsForbidden(err) {
			if strings.Contains(err.Error(), "exceeded quota") {
				logger.Warnf(ctx, "Failed to launch job, resource quota exceeded and the operation is not guarded by back-off. err: %v", err)
				return pluginsCore.PhaseInfoWaitingForResourcesInfo(time.Now(), pluginsCore.DefaultPhaseVersion, fmt.Sprintf("Exceeded resourcequota: %s", err.Error()), nil), nil
			}
			return pluginsCore.PhaseInfoRetryableFailure("RuntimeFailure", err.Error(), nil), nil
		} else if k8serrors.IsBadRequest(err) || k8serrors.IsInvalid(err) {
			logger.Errorf(ctx, "Badly formatted resource for plugin [%s], err %s", executorName, err)
			return pluginsCore.PhaseInfoFailure("BadTaskFormat", err.Error(), nil), nil
		} else if k8serrors.IsRequestEntityTooLargeError(err) {
			logger.Errorf(ctx, "Badly formatted resource for plugin [%s], err %s", executorName, err)
			return pluginsCore.PhaseInfoFailure("EntityTooLarge", err.Error(), nil), nil
		}
		reason := k8serrors.ReasonForError(err)
		logger.Errorf(ctx, "Failed to launch job, system error. err: %v", err)
		return pluginsCore.PhaseInfoUndefined, errors.Wrapf(stdErrors.ErrorCode(reason), err, "failed to create resource")
	}

	return pluginsCore.PhaseInfoQueued(time.Now(), pluginsCore.DefaultPhaseVersion, "task submitted to K8s"), nil
}

// finalizeSubtask performs operations to complete the k8s pod defined by the SubTaskExecutionContext
// and Config. These may include removing finalizers and deleting the k8s resource.
func finalizeSubtask(ctx context.Context, stCtx SubTaskExecutionContext, cfg *Config, kubeClient pluginsCore.KubeClient) error {
	errs := stdErrors.ErrorCollection{}
	var pod *v1.Pod
	var nsName k8stypes.NamespacedName
	k8sPluginCfg := config.GetK8sPluginConfig()
	if k8sPluginCfg.InjectFinalizer || k8sPluginCfg.DeleteResourceOnFinalize {
		o, err := podPlugin.DefaultPodPlugin.BuildIdentityResource(ctx, stCtx.TaskExecutionMetadata())
		if err != nil {
			// This will recurrent, so we will skip further finalize
			logger.Errorf(ctx, "Failed to build the Resource with name: %v. Error: %v, when finalizing.", stCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), err)
			return nil
		}

		pod = o.(*v1.Pod)

		addMetadata(stCtx, cfg, config.GetK8sPluginConfig(), pod)
		nsName = k8stypes.NamespacedName{Namespace: pod.GetNamespace(), Name: pod.GetName()}
	}

	// In InjectFinalizer is on, it means we may have added the finalizers when we launched this resource. Attempt to
	// clear them to allow the object to be deleted/garbage collected. If InjectFinalizer was turned on (through config)
	// after the resource was created, we will not find any finalizers to clear and the object may have already been
	// deleted at this point. Therefore, account for these cases and do not consider them errors.
	if k8sPluginCfg.InjectFinalizer {
		// Attempt to get resource from informer cache, if not found, retrieve it from API server.
		if err := kubeClient.GetClient().Get(ctx, nsName, pod); err != nil {
			if isK8sObjectNotExists(err) {
				return nil
			}
			// This happens sometimes because a node gets removed and K8s deletes the pod. This will result in a
			// Pod does not exist error. This should be retried using the retry policy
			logger.Warningf(ctx, "Failed in finalizing get Resource with name: %v. Error: %v", nsName, err)
			return err
		}

		// This must happen after sending admin event. It's safe against partial failures because if the event failed, we will
		// simply retry in the next round. If the event succeeded but this failed, we will try again the next round to send
		// the same event (idempotent) and then come here again...
		err := clearFinalizers(ctx, pod, kubeClient)
		if err != nil {
			errs.Append(err)
		}
	}

	// If we should delete the resource when finalize is called, do a best effort delete.
	if k8sPluginCfg.DeleteResourceOnFinalize {
		// Attempt to delete resource, if not found, return success.
		if err := kubeClient.GetClient().Delete(ctx, pod); err != nil {
			if isK8sObjectNotExists(err) {
				return errs.ErrorOrDefault()
			}

			// This happens sometimes because a node gets removed and K8s deletes the pod. This will result in a
			// Pod does not exist error. This should be retried using the retry policy
			logger.Warningf(ctx, "Failed in finalizing. Failed to delete Resource with name: %v. Error: %v", nsName, err)
			errs.Append(fmt.Errorf("finalize: failed to delete resource with name [%v]. Error: %w", nsName, err))
		}
	}

	return errs.ErrorOrDefault()
}

// getSubtaskPhaseInfo returns the PhaseInfo describing an existing k8s resource which is defined
// by the SubTaskExecutionContext and Config.
func getSubtaskPhaseInfo(ctx context.Context, stCtx SubTaskExecutionContext, cfg *Config, kubeClient pluginsCore.KubeClient, logPlugin tasklog.Plugin) (pluginsCore.PhaseInfo, error) {
	o, err := podPlugin.DefaultPodPlugin.BuildIdentityResource(ctx, stCtx.TaskExecutionMetadata())
	if err != nil {
		logger.Errorf(ctx, "Failed to build the Resource with name: %v. Error: %v", stCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), err)
		return pluginsCore.PhaseInfoFailure("BadTaskDefinition", fmt.Sprintf("Failed to build resource, caused by: %s", err.Error()), nil), nil
	}

	pod := o.(*v1.Pod)
	addMetadata(stCtx, cfg, config.GetK8sPluginConfig(), pod)

	// Attempt to get resource from informer cache, if not found, retrieve it from API server.
	nsName := k8stypes.NamespacedName{Name: pod.GetName(), Namespace: pod.GetNamespace()}
	if err := kubeClient.GetClient().Get(ctx, nsName, pod); err != nil {
		if isK8sObjectNotExists(err) {
			// This happens sometimes because a node gets removed and K8s deletes the pod. This will result in a
			// Pod does not exist error. This should be retried using the retry policy
			logger.Warnf(ctx, "Failed to find the Resource with name: %v. Error: %v", nsName, err)
			failureReason := fmt.Sprintf("resource not found, name [%s]. reason: %s", nsName.String(), err.Error())
			return pluginsCore.PhaseInfoSystemRetryableFailure("ResourceDeletedExternally", failureReason, nil), nil
		}

		logger.Warnf(ctx, "Failed to retrieve Resource Details with name: %v. Error: %v", nsName, err)
		return pluginsCore.PhaseInfoUndefined, err
	}

	stID, _ := stCtx.TaskExecutionMetadata().GetTaskExecutionID().(SubTaskExecutionID)
	phaseInfo, err := podPlugin.DefaultPodPlugin.GetTaskPhaseWithLogs(ctx, stCtx, pod, logPlugin, stID.GetLogSuffix(), stID.TemplateVarsByScheme())
	if err != nil {
		logger.Warnf(ctx, "failed to check status of resource in plugin [%s], with error: %s", executorName, err.Error())
		return pluginsCore.PhaseInfoUndefined, err
	}

	if !phaseInfo.Phase().IsTerminal() && o.GetDeletionTimestamp() != nil {
		// If the object has been deleted, that is, it has a deletion timestamp, but is not in a terminal state, we should
		// mark the task as a retryable failure.  We've seen this happen when a kubelet disappears - all pods running on
		// the node are marked with a deletionTimestamp, but our finalizers prevent the pod from being deleted.
		// This can also happen when a user deletes a Pod directly.
		failureReason := fmt.Sprintf("object [%s] terminated in the background, manually", nsName.String())
		return pluginsCore.PhaseInfoSystemRetryableFailure("UnexpectedObjectDeletion", failureReason, nil), nil
	}

	return phaseInfo, err
}

// getTaskContainerIndex returns the index of the primary container in a k8s pod.
func getTaskContainerIndex(pod *v1.Pod) (int, error) {
	primaryContainerName, ok := pod.Annotations[flytek8s.PrimaryContainerKey]
	// For tasks with a Container target, we only ever build one container as part of the pod
	if !ok {
		if len(pod.Spec.Containers) == 1 {
			return 0, nil
		}
		// For tasks with a K8sPod task target, they may produce multiple containers but at least one must be the designated primary.
		return -1, stdErrors.Errorf(ErrBuildPodTemplate, "Expected a specified primary container key when building an array job with a K8sPod spec target")

	}

	for idx, container := range pod.Spec.Containers {
		if container.Name == primaryContainerName {
			return idx, nil
		}
	}
	return -1, stdErrors.Errorf(ErrBuildPodTemplate, "Couldn't find any container matching the primary container key when building an array job with a K8sPod spec target")
}

// isK8sObjectNotExists returns true if the error is one which describes a non existent k8s object.
func isK8sObjectNotExists(err error) bool {
	return k8serrors.IsNotFound(err) || k8serrors.IsGone(err) || k8serrors.IsResourceExpired(err)
}
