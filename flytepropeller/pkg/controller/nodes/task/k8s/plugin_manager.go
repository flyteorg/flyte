package k8s

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/time/rate"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/errors"
	pluginsCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/k8s"
	pluginsUtils "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils"
	compiler "github.com/flyteorg/flyte/flytepropeller/pkg/compiler/transformers/k8s"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task/backoff"
	nodeTaskConfig "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task/config"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	stdErrors "github.com/flyteorg/flyte/flytestdlib/errors"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
)

const finalizer = "flyte/flytek8s"

const pluginStateVersion = 1

type PluginPhase uint8

const (
	PluginPhaseNotStarted PluginPhase = iota
	PluginPhaseAllocationTokenAcquired
	PluginPhaseStarted
)

type PluginState struct {
	Phase           PluginPhase
	K8sPluginState  k8s.PluginState
	LastEventUpdate time.Time
}

type PluginMetrics struct {
	Scope           promutils.Scope
	GetCacheMiss    labeled.StopWatch
	GetCacheHit     labeled.StopWatch
	GetAPILatency   labeled.StopWatch
	ResourceDeleted labeled.Counter
}

func newPluginMetrics(s promutils.Scope) PluginMetrics {
	return PluginMetrics{
		Scope: s,
		GetCacheMiss: labeled.NewStopWatch("get_cache_miss", "Cache miss on get resource calls.",
			time.Millisecond, s),
		GetCacheHit: labeled.NewStopWatch("get_cache_hit", "Cache miss on get resource calls.",
			time.Millisecond, s),
		GetAPILatency: labeled.NewStopWatch("get_api", "Latency for APIServer Get calls.",
			time.Millisecond, s),
		ResourceDeleted: labeled.NewCounter("pods_deleted", "Counts how many times CheckTaskStatus is"+
			" called with a deleted resource.", s),
	}
}

func isK8sObjectNotExists(err error) bool {
	return k8serrors.IsNotFound(err) || k8serrors.IsGone(err) || k8serrors.IsResourceExpired(err)
}

// A generic Plugin for managing k8s-resources. Plugin writers wishing to use K8s resource can use the simplified api specified in
// pluginmachinery.core
type PluginManager struct {
	id              string
	plugin          k8s.Plugin
	resourceToWatch runtime.Object
	kubeClient      pluginsCore.KubeClient
	metrics         PluginMetrics
	// Per namespace-resource
	backOffController         *backoff.Controller
	resourceLevelMonitor      *ResourceLevelMonitor
	eventWatcher              EventWatcher
	updateBaseBackoffDuration int
	updateBackoffRetries      int
}

func (e *PluginManager) addObjectMetadata(taskCtx pluginsCore.TaskExecutionMetadata, o client.Object, cfg *config.K8sPluginConfig) {
	o.SetNamespace(taskCtx.GetNamespace())
	o.SetAnnotations(pluginsUtils.UnionMaps(cfg.DefaultAnnotations, o.GetAnnotations(), pluginsUtils.CopyMap(taskCtx.GetAnnotations())))
	o.SetLabels(pluginsUtils.UnionMaps(cfg.DefaultLabels, o.GetLabels(), pluginsUtils.CopyMap(taskCtx.GetLabels())))
	o.SetName(taskCtx.GetTaskExecutionID().GetGeneratedName())

	if !e.plugin.GetProperties().DisableInjectOwnerReferences {
		o.SetOwnerReferences([]metav1.OwnerReference{taskCtx.GetOwnerReference()})
	}

	if cfg.InjectFinalizer && !e.plugin.GetProperties().DisableInjectFinalizer {
		f := append(o.GetFinalizers(), finalizer)
		o.SetFinalizers(f)
	}

	if errs := validation.IsDNS1123Subdomain(o.GetName()); len(errs) > 0 {
		o.SetName(pluginsUtils.ConvertToDNS1123SubdomainCompatibleString(o.GetName()))
	}
}

func (e *PluginManager) GetProperties() pluginsCore.PluginProperties {
	props := e.plugin.GetProperties()
	return pluginsCore.PluginProperties{
		GeneratedNameMaxLength: props.GeneratedNameMaxLength,
	}
}

func (e *PluginManager) GetID() string {
	return e.id
}

func (e *PluginManager) getPodEffectiveResourceLimits(ctx context.Context, pod *v1.Pod) v1.ResourceList {
	podRequestedResources := make(v1.ResourceList)
	initContainersRequestedResources := make(v1.ResourceList)
	containersRequestedResources := make(v1.ResourceList)

	// Collect the resource requests from all the containers in the pod whose creation is to be attempted
	// to decide whether we should try the pod creation during the back off period

	// Calculating the effective init resource limits based on the official definition:
	// https://kubernetes.io/docs/concepts/workloads/pods/init-containers/#resources
	// "The highest of any particular resource request or limit defined on all init containers is the effective init request/limit"
	for _, initContainer := range pod.Spec.InitContainers {
		for r, q := range initContainer.Resources.Limits {
			if currentQuantity, found := initContainersRequestedResources[r]; !found || q.Cmp(currentQuantity) > 0 {
				initContainersRequestedResources[r] = q
			}
		}
	}

	for _, container := range pod.Spec.Containers {
		for k, v := range container.Resources.Limits {
			quantity := containersRequestedResources[k]
			quantity.Add(v)
			containersRequestedResources[k] = quantity
		}
	}

	for k, v := range initContainersRequestedResources {
		podRequestedResources[k] = v
	}

	// https://kubernetes.io/docs/concepts/workloads/pods/init-containers/#resources
	// "The Pod’s effective request/limit for a resource is the higher of:
	// - the sum of all app containers request/limit for a resource
	// - the effective init request/limit for a resource"
	for k, qC := range containersRequestedResources {
		if qI, found := podRequestedResources[k]; !found || qC.Cmp(qI) > 0 {
			podRequestedResources[k] = qC
		}
	}

	formattedResources := make([]string, 0, len(podRequestedResources))
	for resourceName, quantity := range podRequestedResources {
		formattedResources = append(formattedResources, fmt.Sprintf("{[%v]: [%v]}", resourceName, quantity.String()))
	}

	logger.Infof(ctx, "The resource requirement for creating Pod [%v/%v] is [%+v]\n",
		pod.Namespace, pod.Name, formattedResources)

	return podRequestedResources
}

func (e *PluginManager) launchResource(ctx context.Context, tCtx pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error) {
	tmpl, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return pluginsCore.Transition{}, err
	}

	k8sTaskCtxMetadata, err := newTaskExecutionMetadata(tCtx.TaskExecutionMetadata(), tmpl)
	if err != nil {
		return pluginsCore.Transition{}, err
	}

	k8sTaskCtx := newTaskExecutionContext(tCtx, k8sTaskCtxMetadata)

	o, err := e.plugin.BuildResource(ctx, k8sTaskCtx)
	if err != nil {
		return pluginsCore.UnknownTransition, err
	}

	e.addObjectMetadata(k8sTaskCtxMetadata, o, config.GetK8sPluginConfig())
	logger.Infof(ctx, "Creating Object: Type:[%v], Object:[%v/%v]", o.GetObjectKind().GroupVersionKind(), o.GetNamespace(), o.GetName())

	key := backoff.ComposeResourceKey(o)

	pod, casted := o.(*v1.Pod)
	if e.backOffController != nil && casted {
		podRequestedResources := e.getPodEffectiveResourceLimits(ctx, pod)

		cfg := nodeTaskConfig.GetConfig()
		backOffHandler := e.backOffController.GetOrCreateHandler(ctx, key, cfg.BackOffConfig.BaseSecond, cfg.BackOffConfig.MaxDuration.Duration)

		err = backOffHandler.Handle(ctx, func() error {
			return e.kubeClient.GetClient().Create(ctx, o)
		}, podRequestedResources)
	} else {
		err = e.kubeClient.GetClient().Create(ctx, o)
	}

	if err != nil && !k8serrors.IsAlreadyExists(err) {
		if backoff.IsResourceQuotaExceeded(err) && !backoff.IsResourceRequestsEligible(err) {
			// if task resources exceed resource quotas then permanently fail because the task will
			// be stuck waiting for resources until the `node-active-deadline` terminates the node.
			logger.Errorf(ctx, "task resource requests exceed k8s resource limits. err: %v", err)
			return pluginsCore.DoTransition(pluginsCore.PhaseInfoFailure("ResourceRequestsExceedLimits",
				fmt.Sprintf("requested resources exceed limits: %v", err.Error()), nil)), nil
		} else if stdErrors.IsCausedBy(err, errors.BackOffError) {
			logger.Warnf(ctx, "Failed to launch job, resource quota exceeded. err: %v", err)
			return pluginsCore.DoTransition(pluginsCore.PhaseInfoWaitingForResourcesInfo(time.Now(), pluginsCore.DefaultPhaseVersion, fmt.Sprintf("Exceeded resourcequota: %s", err.Error()), nil)), nil
		} else if e.backOffController == nil && backoff.IsResourceQuotaExceeded(err) {
			logger.Warnf(ctx, "Failed to launch job, resource quota exceeded and the operation is not guarded by back-off. err: %v", err)
			return pluginsCore.DoTransition(pluginsCore.PhaseInfoWaitingForResourcesInfo(time.Now(), pluginsCore.DefaultPhaseVersion, fmt.Sprintf("Exceeded resourcequota: %s", err.Error()), nil)), nil
		} else if k8serrors.IsForbidden(err) {
			return pluginsCore.DoTransition(pluginsCore.PhaseInfoRetryableFailure("RuntimeFailure", err.Error(), nil)), nil
		} else if k8serrors.IsBadRequest(err) || k8serrors.IsInvalid(err) {
			logger.Errorf(ctx, "Badly formatted resource for plugin [%s], err %s", e.id, err)
			// return pluginsCore.DoTransition(pluginsCore.PhaseInfoFailure("BadTaskFormat", err.Error(), nil)), nil
		} else if k8serrors.IsRequestEntityTooLargeError(err) {
			logger.Errorf(ctx, "Badly formatted resource for plugin [%s], err %s", e.id, err)
			return pluginsCore.DoTransition(pluginsCore.PhaseInfoFailure("EntityTooLarge", err.Error(), nil)), nil
		}
		reason := k8serrors.ReasonForError(err)
		logger.Errorf(ctx, "Failed to launch job, system error. err: %v", err)
		return pluginsCore.UnknownTransition, errors.Wrapf(stdErrors.ErrorCode(reason), err, "failed to create resource")
	}

	return pluginsCore.DoTransition(pluginsCore.PhaseInfoQueued(time.Now(), pluginsCore.DefaultPhaseVersion, "task submitted to K8s")), nil
}

func (e *PluginManager) getResource(ctx context.Context, tCtx pluginsCore.TaskExecutionContext) (client.Object, error) {
	o, err := e.plugin.BuildIdentityResource(ctx, tCtx.TaskExecutionMetadata())
	if err != nil {
		logger.Errorf(ctx, "Failed to build the Resource with name: %v. Error: %v", tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), err)
		return nil, err
	}
	e.addObjectMetadata(tCtx.TaskExecutionMetadata(), o, config.GetK8sPluginConfig())
	return o, nil
}

func (e *PluginManager) checkResourcePhase(ctx context.Context, tCtx pluginsCore.TaskExecutionContext, o client.Object, k8sPluginState *k8s.PluginState) (pluginsCore.Transition, error) {
	nsName := k8stypes.NamespacedName{Namespace: o.GetNamespace(), Name: o.GetName()}
	// Attempt to get resource from informer cache, if not found, retrieve it from API server.
	if err := e.kubeClient.GetClient().Get(ctx, nsName, o); err != nil {
		if isK8sObjectNotExists(err) {
			// This happens sometimes because a node gets removed and K8s deletes the pod. This will result in a
			// Pod does not exist error. This should be retried using the retry policy
			logger.Warningf(ctx, "Failed to find the Resource with name: %v. Error: %v", nsName, err)
			failureReason := fmt.Sprintf("resource not found, name [%s]. reason: %s", nsName.String(), err.Error())
			return pluginsCore.DoTransition(pluginsCore.PhaseInfoSystemRetryableFailure("ResourceDeletedExternally", failureReason, nil)), nil
		}

		logger.Warningf(ctx, "Failed to retrieve Resource Details with name: %v. Error: %v", nsName, err)
		return pluginsCore.UnknownTransition, err
	}
	if o.GetDeletionTimestamp() != nil {
		e.metrics.ResourceDeleted.Inc(ctx)
	}

	pCtx := newPluginContext(tCtx, k8sPluginState)
	p, err := e.plugin.GetTaskPhase(ctx, pCtx, o)
	if err != nil {
		logger.Warnf(ctx, "failed to check status of resource in plugin [%s], with error: %s", e.GetID(), err.Error())
		return pluginsCore.UnknownTransition, err
	}

	if p.Phase() == pluginsCore.PhaseSuccess {
		var opReader io.OutputReader
		if pCtx.ow == nil {
			logger.Infof(ctx, "Plugin [%s] returned no outputReader, assuming file based outputs", e.id)
			opReader = ioutils.NewRemoteFileOutputReader(ctx, tCtx.DataStore(), tCtx.OutputWriter(), 0)
		} else {
			logger.Infof(ctx, "Plugin [%s] returned outputReader", e.id)
			opReader = pCtx.ow.GetReader()
		}
		err := tCtx.OutputWriter().Put(ctx, opReader)
		if err != nil {
			return pluginsCore.UnknownTransition, err
		}

		return pluginsCore.DoTransition(p), nil
	}

	if !p.Phase().IsTerminal() && o.GetDeletionTimestamp() != nil {
		// If the object has been deleted, that is, it has a deletion timestamp, but is not in a terminal state, we should
		// mark the task as a retryable failure.  We've seen this happen when a kubelet disappears - all pods running on
		// the node are marked with a deletionTimestamp, but our finalizers prevent the pod from being deleted.
		// This can also happen when a user deletes a Pod directly.
		failureReason := fmt.Sprintf("object [%s] terminated in the background, manually", nsName.String())
		return pluginsCore.DoTransition(pluginsCore.PhaseInfoSystemRetryableFailure("UnexpectedObjectDeletion", failureReason, nil)), nil
	}

	return pluginsCore.DoTransition(p), nil
}

func (e PluginManager) Handle(ctx context.Context, tCtx pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error) {
	// read phase state
	pluginState := PluginState{}
	if v, err := tCtx.PluginStateReader().Get(&pluginState); err != nil {
		if v != pluginStateVersion {
			return pluginsCore.DoTransition(pluginsCore.PhaseInfoRetryableFailure(errors.CorruptedPluginState,
				fmt.Sprintf("plugin state version mismatch expected [%d] got [%d]", pluginStateVersion, v), nil)), nil
		}
		return pluginsCore.UnknownTransition, errors.Wrapf(errors.CorruptedPluginState, err, "Failed to read unmarshal custom state")
	}

	// evaluate plugin
	var err error
	var transition pluginsCore.Transition
	var o client.Object
	pluginPhase := pluginState.Phase
	if pluginState.Phase == PluginPhaseNotStarted {
		transition, err = e.launchResource(ctx, tCtx)
		if err == nil && transition.Info().Phase() == pluginsCore.PhaseQueued {
			pluginPhase = PluginPhaseStarted
		}
	} else {
		o, err = e.getResource(ctx, tCtx)
		if err != nil {
			transition, err = pluginsCore.DoTransition(pluginsCore.PhaseInfoFailure("BadTaskDefinition",
				fmt.Sprintf("Failed to build resource, caused by: %s", err.Error()), nil)), nil
		} else {
			transition, err = e.checkResourcePhase(ctx, tCtx, o, &pluginState.K8sPluginState)
		}
	}

	if err != nil {
		return transition, err
	}

	// Add events since last update
	version := transition.Info().Version()
	lastEventUpdate := pluginState.LastEventUpdate
	if e.eventWatcher != nil && o != nil {
		nsName := k8stypes.NamespacedName{Namespace: o.GetNamespace(), Name: o.GetName()}
		recentEvents := e.eventWatcher.List(nsName, lastEventUpdate)
		if len(recentEvents) > 0 {
			taskInfo := transition.Info().Info()
			taskInfo.AdditionalReasons = make([]pluginsCore.ReasonInfo, 0, len(recentEvents))
			for _, event := range recentEvents {
				taskInfo.AdditionalReasons = append(taskInfo.AdditionalReasons,
					pluginsCore.ReasonInfo{Reason: event.Note, OccurredAt: &event.CreatedAt})
				lastEventUpdate = event.CreatedAt
			}
			// Bump the version to ensure newly added events are picked up
			version++
		}
	}

	// persist any changes in phase state
	newPluginState := PluginState{
		Phase: pluginPhase,
		K8sPluginState: k8s.PluginState{
			Phase:        transition.Info().Phase(),
			PhaseVersion: version,
			Reason:       transition.Info().Reason(),
		},
		LastEventUpdate: lastEventUpdate,
	}
	if pluginState != newPluginState {
		if err := tCtx.PluginStateWriter().Put(pluginStateVersion, &newPluginState); err != nil {
			return pluginsCore.UnknownTransition, err
		}
	}

	return transition, nil
}

func (e PluginManager) Abort(ctx context.Context, tCtx pluginsCore.TaskExecutionContext) error {
	logger.Infof(ctx, "KillTask invoked. We will attempt to delete object [%v].",
		tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName())

	o, err := e.getResource(ctx, tCtx)
	if err != nil {
		logger.Errorf(ctx, "%v", err)
		return nil
	}

	deleteResource := true
	abortOverride, hasAbortOverride := e.plugin.(k8s.PluginAbortOverride)

	resourceToFinalize := o
	var behavior k8s.AbortBehavior

	if hasAbortOverride {
		behavior, err = abortOverride.OnAbort(ctx, tCtx, o)
		deleteResource = err == nil && behavior.DeleteResource
		if err == nil && behavior.Resource != nil {
			resourceToFinalize = behavior.Resource
		}
	}

	if err != nil {
	} else if deleteResource {
		err = e.kubeClient.GetClient().Delete(ctx, resourceToFinalize)
	} else {
		if behavior.Patch != nil && behavior.Update == nil {
			err = e.kubeClient.GetClient().Patch(ctx, resourceToFinalize, behavior.Patch.Patch, behavior.Patch.Options...)
		} else if behavior.Patch == nil && behavior.Update != nil {
			err = e.kubeClient.GetClient().Update(ctx, resourceToFinalize, behavior.Update.Options...)
		} else {
			err = errors.Errorf(errors.RuntimeFailure, "AbortBehavior for resource %v must specify either a Patch and an Update operation if Delete is set to false. Only one can be supplied.", resourceToFinalize.GetName())
		}
		if behavior.DeleteOnErr && err != nil {
			logger.Warningf(ctx, "Failed to apply AbortBehavior for resource %v with error %v. Will attempt to delete resource.", resourceToFinalize.GetName(), err)
			err = e.kubeClient.GetClient().Delete(ctx, resourceToFinalize)
		}
	}

	if err != nil && !isK8sObjectNotExists(err) {
		logger.Warningf(ctx, "Failed to clear finalizers for Resource with name: %v/%v. Error: %v",
			resourceToFinalize.GetNamespace(), resourceToFinalize.GetName(), err)
		return err
	}

	return nil
}

func (e *PluginManager) clearFinalizers(ctx context.Context, o client.Object) error {
	if len(o.GetFinalizers()) > 0 {
		o.SetFinalizers([]string{})
		err := e.kubeClient.GetClient().Update(ctx, o)
		if err != nil && !isK8sObjectNotExists(err) {
			logger.Warningf(ctx, "Failed to clear finalizers for Resource with name: %v/%v. Error: %v",
				o.GetNamespace(), o.GetName(), err)
			return err
		}
	} else {
		logger.Debugf(ctx, "Finalizers are already empty for Resource with name: %v/%v",
			o.GetNamespace(), o.GetName())
	}
	return nil
}

func (e *PluginManager) Finalize(ctx context.Context, tCtx pluginsCore.TaskExecutionContext) (err error) {
	errs := stdErrors.ErrorCollection{}
	var nsName k8stypes.NamespacedName
	cfg := config.GetK8sPluginConfig()

	o, err := e.getResource(ctx, tCtx)
	if err != nil {
		logger.Errorf(ctx, "%v", err)
		return nil
	}
	nsName = k8stypes.NamespacedName{Namespace: o.GetNamespace(), Name: o.GetName()}

	retryBackoff := wait.Backoff{
		Duration: time.Duration(e.updateBaseBackoffDuration) * time.Millisecond,
		Factor:   2.0,
		Jitter:   0.1,
		Steps:    e.updateBackoffRetries,
	}

	// Attempt to cleanup finalizers so that the object may be deleted/garbage collected. We try to clear them for all
	// objects, regardless of whether or not InjectFinalizer is configured to handle all cases where InjectFinalizer is
	// enabled/disabled during object execution.
	var lastErr error
	_ = wait.ExponentialBackoff(retryBackoff, func() (bool, error) {
		lastErr = nil
		if err := e.kubeClient.GetClient().Get(ctx, nsName, o); err != nil {
			if isK8sObjectNotExists(err) {
				return true, nil
			}
			lastErr = err
			// This happens sometimes because a node gets removed and K8s deletes the pod. This will result in a
			// Pod does not exist error. This should be retried using the retry policy
			logger.Warningf(ctx, "Failed in finalizing get Resource with name: %v. Error: %v", nsName, err)
			return true, err
		}

		// This must happen after sending admin event. It's safe against partial failures because if the event failed, we will
		// simply retry in the next round. If the event succeeded but this failed, we will try again the next round to send
		// the same event (idempotent) and then come here again...
		if err := e.clearFinalizers(ctx, o); err != nil {
			lastErr = err
			// retry is if there is a conflict in case the informer cache is out of sync
			if k8serrors.IsConflict(err) {
				logger.Warningf(ctx, "Failed to clear finalizers for Resource with name: %v. Error: %v. Retrying..", nsName, err)
				return false, nil
			}
			logger.Warningf(ctx, "Failed to clear finalizers for Resource with name: %v. Error: %v", nsName, err)
			return true, err
		}
		return true, nil
	})

	if lastErr != nil {
		errs.Append(lastErr)
	}

	// If we should delete the resource when finalize is called, do a best effort delete.
	if cfg.DeleteResourceOnFinalize && !e.plugin.GetProperties().DisableDeleteResourceOnFinalize {
		// Attempt to delete resource, if not found, return success.
		if err := e.kubeClient.GetClient().Delete(ctx, o); err != nil {
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

func NewPluginManagerWithBackOff(ctx context.Context, iCtx pluginsCore.SetupContext, entry k8s.PluginEntry, backOffController *backoff.Controller,
	monitorIndex *ResourceMonitorIndex, kubeClientset kubernetes.Interface) (*PluginManager, error) {

	mgr, err := NewPluginManager(ctx, iCtx, entry, monitorIndex, kubeClientset)
	if err == nil {
		mgr.backOffController = backOffController
	}
	return mgr, err
}

// Creates a K8s generic task executor. This provides an easier way to build task executors that create K8s resources.
func NewPluginManager(ctx context.Context, iCtx pluginsCore.SetupContext, entry k8s.PluginEntry,
	monitorIndex *ResourceMonitorIndex, kubeClientset kubernetes.Interface) (*PluginManager, error) {

	if iCtx.EnqueueOwner() == nil {
		return nil, errors.Errorf(errors.PluginInitializationFailed, "Failed to initialize plugin, enqueue Owner cannot be nil or empty.")
	}

	kubeClient := iCtx.KubeClient()
	if entry.CustomKubeClient != nil {
		kc, err := entry.CustomKubeClient(ctx)
		if err != nil {
			return nil, err
		}

		if kc != nil {
			kubeClient = kc
		}
	}

	if kubeClient == nil {
		return nil, errors.Errorf(errors.PluginInitializationFailed, "Failed to initialize K8sResource Plugin, Kubeclient cannot be nil!")
	}

	logger.Infof(ctx, "Initializing K8s plugin [%s]", entry.ID)
	src := source.Kind(iCtx.KubeClient().GetCache(), entry.ResourceToWatch)

	workflowParentPredicate := func(o metav1.Object) bool {
		if entry.Plugin.GetProperties().DisableInjectOwnerReferences {
			return true
		}

		ownerReference := metav1.GetControllerOf(o)
		if ownerReference != nil {
			if ownerReference.Kind == iCtx.OwnerKind() {
				return true
			}
		}

		return false
	}

	metricsScope := iCtx.MetricsScope().NewSubScope(entry.ID)
	updateCount := labeled.NewCounter("informer_update", "Update events from informer", metricsScope)
	droppedUpdateCount := labeled.NewCounter("informer_update_dropped", "Update events from informer that have the same resource version", metricsScope)
	genericCount := labeled.NewCounter("informer_generic", "Generic events from informer", metricsScope)

	enqueueOwner := iCtx.EnqueueOwner()
	err := src.Start(
		ctx,
		// Handlers
		handler.Funcs{
			CreateFunc: func(ctx context.Context, evt event.CreateEvent, q2 workqueue.RateLimitingInterface) {
				logger.Debugf(context.Background(), "Create received for %s, ignoring.", evt.Object.GetName())
			},
			UpdateFunc: func(ctx context.Context, evt event.UpdateEvent, q2 workqueue.RateLimitingInterface) {
				if evt.ObjectNew == nil {
					logger.Warn(context.Background(), "Received an Update event with nil MetaNew.")
				} else if evt.ObjectOld == nil || evt.ObjectOld.GetResourceVersion() != evt.ObjectNew.GetResourceVersion() {
					// attempt to enqueue this tasks owner by retrieving the workfowID from the resource labels
					newCtx := contextutils.WithNamespace(context.Background(), evt.ObjectNew.GetNamespace())

					workflowID, exists := evt.ObjectNew.GetLabels()[compiler.ExecutionIDLabel]
					if exists {
						logger.Debugf(ctx, "Enqueueing owner for updated object [%v/%v]", evt.ObjectNew.GetNamespace(), evt.ObjectNew.GetName())
						namespacedName := k8stypes.NamespacedName{
							Name:      workflowID,
							Namespace: evt.ObjectNew.GetNamespace(),
						}

						if err := enqueueOwner(namespacedName); err != nil {
							logger.Warnf(context.Background(), "Failed to handle Update event for object [%v]", namespacedName)
						}
						updateCount.Inc(newCtx)
					}
				} else {
					newCtx := contextutils.WithNamespace(context.Background(), evt.ObjectNew.GetNamespace())
					droppedUpdateCount.Inc(newCtx)
				}
			},
			DeleteFunc: func(ctx context.Context, evt event.DeleteEvent, q2 workqueue.RateLimitingInterface) {
				logger.Debugf(context.Background(), "Delete received for %s, ignoring.", evt.Object.GetName())
			},
			GenericFunc: func(ctx context.Context, evt event.GenericEvent, q2 workqueue.RateLimitingInterface) {
				logger.Debugf(context.Background(), "Generic received for %s, ignoring.", evt.Object.GetName())
				genericCount.Inc(ctx)
			},
		},
		// Queue - configured for high throughput so we very infrequently rate limit node updates
		workqueue.NewNamedRateLimitingQueue(&workqueue.BucketRateLimiter{
			Limiter: rate.NewLimiter(rate.Limit(10000), 10000),
		}, entry.ResourceToWatch.GetObjectKind().GroupVersionKind().Kind),
		// Predicates
		predicate.Funcs{
			CreateFunc: func(createEvent event.CreateEvent) bool {
				return false
			},
			UpdateFunc: func(updateEvent event.UpdateEvent) bool {
				// TODO we should filter out events in case there are no updates observed between the old and new?
				return workflowParentPredicate(updateEvent.ObjectNew)
			},
			DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
				return false
			},
			GenericFunc: func(genericEvent event.GenericEvent) bool {
				return workflowParentPredicate(genericEvent.Object)
			},
		})

	if err != nil {
		return nil, err
	}

	gvk, err := getPluginGvk(entry.ResourceToWatch)
	if err != nil {
		return nil, err
	}

	k8sConfig := config.GetK8sPluginConfig()
	var eventWatcher EventWatcher
	if k8sConfig.SendObjectEvents {
		eventWatcher, err = NewEventWatcher(ctx, gvk, kubeClientset)
		if err != nil {
			return nil, err
		}
	}

	// Construct the collector that will emit a gauge indicating current levels of the resource that this K8s plugin operates on
	rm := monitorIndex.GetOrCreateResourceLevelMonitor(ctx, metricsScope, kubeClient.GetCache(), gvk)

	// Start the poller and gauge emitter
	rm.RunCollectorOnce(ctx)

	return &PluginManager{
		id:                        entry.ID,
		plugin:                    entry.Plugin,
		resourceToWatch:           entry.ResourceToWatch,
		metrics:                   newPluginMetrics(metricsScope),
		kubeClient:                kubeClient,
		resourceLevelMonitor:      rm,
		eventWatcher:              eventWatcher,
		updateBaseBackoffDuration: k8sConfig.UpdateBaseBackoffDuration,
		updateBackoffRetries:      k8sConfig.UpdateBackoffRetries,
	}, nil
}

func getPluginGvk(resourceToWatch runtime.Object) (schema.GroupVersionKind, error) {
	kinds, _, err := scheme.Scheme.ObjectKinds(resourceToWatch)
	if err != nil && len(kinds) == 0 {
		return schema.GroupVersionKind{}, errors.Errorf(errors.PluginInitializationFailed, "No kind in schema for %v", resourceToWatch)
	}
	return kinds[0], nil
}
