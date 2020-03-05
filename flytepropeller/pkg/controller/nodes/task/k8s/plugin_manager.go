package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lyft/flytepropeller/pkg/controller/nodes/task/backoff"
	v1 "k8s.io/api/core/v1"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/lyft/flytestdlib/contextutils"
	stdErrors "github.com/lyft/flytestdlib/errors"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/promutils/labeled"

	pluginsCore "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/utils"
	k8stypes "k8s.io/apimachinery/pkg/types"

	"github.com/lyft/flytestdlib/logger"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/lyft/flyteplugins/go/tasks/errors"

	nodeTaskConfig "github.com/lyft/flytepropeller/pkg/controller/nodes/task/config"
	"sigs.k8s.io/controller-runtime/pkg/handler"
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
	Phase PluginPhase
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

func AddObjectMetadata(taskCtx pluginsCore.TaskExecutionMetadata, o k8s.Resource, cfg *config.K8sPluginConfig) {
	o.SetNamespace(taskCtx.GetNamespace())
	o.SetAnnotations(utils.UnionMaps(cfg.DefaultAnnotations, o.GetAnnotations(), utils.CopyMap(taskCtx.GetAnnotations())))
	o.SetLabels(utils.UnionMaps(o.GetLabels(), utils.CopyMap(taskCtx.GetLabels()), cfg.DefaultLabels))
	o.SetOwnerReferences([]metav1.OwnerReference{taskCtx.GetOwnerReference()})
	o.SetName(taskCtx.GetTaskExecutionID().GetGeneratedName())
	if cfg.InjectFinalizer {
		f := append(o.GetFinalizers(), finalizer)
		o.SetFinalizers(f)
	}
}

func IsK8sObjectNotExists(err error) bool {
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
	backOffController *backoff.Controller
}

func (e *PluginManager) GetProperties() pluginsCore.PluginProperties {
	return pluginsCore.PluginProperties{}
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

	logger.Infof(ctx, "The resource requirement for creating Pod [%v/%v] is [%v]\n",
		pod.Namespace, pod.Name, podRequestedResources)

	return podRequestedResources
}

func (e *PluginManager) LaunchResource(ctx context.Context, tCtx pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error) {

	o, err := e.plugin.BuildResource(ctx, tCtx)
	if err != nil {
		return pluginsCore.UnknownTransition, err
	}

	AddObjectMetadata(tCtx.TaskExecutionMetadata(), o, config.GetK8sPluginConfig())
	logger.Infof(ctx, "Creating Object: Type:[%v], Object:[%v/%v]", o.GroupVersionKind(), o.GetNamespace(), o.GetName())

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
		if backoff.IsBackoffError(err) {
			logger.Warnf(ctx, "Failed to launch job, resource quota exceeded. err: %v", err)
			return pluginsCore.DoTransition(pluginsCore.PhaseInfoWaitingForResources(time.Now(), pluginsCore.DefaultPhaseVersion, "failed to launch job, resource quota exceeded.")), nil
		} else if k8serrors.IsForbidden(err) {
			if e.backOffController == nil && strings.Contains(err.Error(), "exceeded quota") {
				logger.Warnf(ctx, "Failed to launch job, resource quota exceeded and the operation is not guarded by back-off. err: %v", err)
				return pluginsCore.DoTransition(pluginsCore.PhaseInfoWaitingForResources(time.Now(), pluginsCore.DefaultPhaseVersion, "failed to launch job, resource quota exceeded.")), nil
			}
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

func (e *PluginManager) CheckResourcePhase(ctx context.Context, tCtx pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error) {

	o, err := e.plugin.BuildIdentityResource(ctx, tCtx.TaskExecutionMetadata())
	if err != nil {
		logger.Errorf(ctx, "Failed to build the Resource with name: %v. Error: %v", tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), err)
		return pluginsCore.DoTransition(pluginsCore.PhaseInfoFailure("BadTaskDefinition", fmt.Sprintf("Failed to build resource, caused by: %s", err.Error()), nil)), nil
	}

	AddObjectMetadata(tCtx.TaskExecutionMetadata(), o, config.GetK8sPluginConfig())
	nsName := k8stypes.NamespacedName{Namespace: o.GetNamespace(), Name: o.GetName()}
	// Attempt to get resource from informer cache, if not found, retrieve it from API server.
	if err := e.kubeClient.GetClient().Get(ctx, nsName, o); err != nil {
		if IsK8sObjectNotExists(err) {
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

	pCtx := newPluginContext(tCtx)
	p, err := e.plugin.GetTaskPhase(ctx, pCtx, o)
	if err != nil {
		logger.Warnf(ctx, "failed to check status of resource in plugin [%s], with error: %s", e.GetID(), err.Error())
		return pluginsCore.UnknownTransition, err
	}

	if p.Phase() == pluginsCore.PhaseSuccess {
		var opReader io.OutputReader
		if pCtx.ow == nil {
			logger.Infof(ctx, "Plugin [%s] returned no outputReader, assuming file based outputs", e.id)
			opReader = ioutils.NewRemoteFileOutputReader(ctx, tCtx.DataStore(), tCtx.OutputWriter(), tCtx.MaxDatasetSizeBytes())
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
	ps := PluginState{}
	if v, err := tCtx.PluginStateReader().Get(&ps); err != nil {
		if v != pluginStateVersion {
			return pluginsCore.DoTransition(pluginsCore.PhaseInfoRetryableFailure(errors.CorruptedPluginState, fmt.Sprintf("plugin state version mismatch expected [%d] got [%d]", pluginStateVersion, v), nil)), nil
		}
		return pluginsCore.UnknownTransition, errors.Wrapf(errors.CorruptedPluginState, err, "Failed to read unmarshal custom state")
	}
	if ps.Phase == PluginPhaseNotStarted {
		t, err := e.LaunchResource(ctx, tCtx)
		if err == nil && t.Info().Phase() == pluginsCore.PhaseQueued {
			if err := tCtx.PluginStateWriter().Put(pluginStateVersion, &PluginState{Phase: PluginPhaseStarted}); err != nil {
				return pluginsCore.UnknownTransition, err
			}
		}
		return t, err
	}
	return e.CheckResourcePhase(ctx, tCtx)
}

func (e PluginManager) Abort(ctx context.Context, tCtx pluginsCore.TaskExecutionContext) error {
	logger.Infof(ctx, "KillTask invoked for %v, nothing to be done.", tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName())
	return nil
}

func (e *PluginManager) ClearFinalizers(ctx context.Context, o k8s.Resource) error {
	if len(o.GetFinalizers()) > 0 {
		o.SetFinalizers([]string{})
		err := e.kubeClient.GetClient().Update(ctx, o)
		if err != nil && !IsK8sObjectNotExists(err) {
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

func (e *PluginManager) Finalize(ctx context.Context, tCtx pluginsCore.TaskExecutionContext) error {
	// If you change InjectFinalizer on the
	if config.GetK8sPluginConfig().InjectFinalizer {
		o, err := e.plugin.BuildIdentityResource(ctx, tCtx.TaskExecutionMetadata())
		if err != nil {
			// This will recurrent, so we will skip further finalize
			logger.Errorf(ctx, "Failed to build the Resource with name: %v. Error: %v, when finalizing.", tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), err)
			return nil
		}

		AddObjectMetadata(tCtx.TaskExecutionMetadata(), o, config.GetK8sPluginConfig())
		nsName := k8stypes.NamespacedName{Namespace: o.GetNamespace(), Name: o.GetName()}
		// Attempt to get resource from informer cache, if not found, retrieve it from API server.
		if err := e.kubeClient.GetClient().Get(ctx, nsName, o); err != nil {
			if IsK8sObjectNotExists(err) {
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
		err = e.ClearFinalizers(ctx, o)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewPluginManagerWithBackOff(ctx context.Context, iCtx pluginsCore.SetupContext, entry k8s.PluginEntry, backOffController *backoff.Controller) (*PluginManager, error) {
	mgr, err := NewPluginManager(ctx, iCtx, entry)
	if err == nil {
		mgr.backOffController = backOffController
	}
	return mgr, err
}

// Creates a K8s generic task executor. This provides an easier way to build task executors that create K8s resources.
func NewPluginManager(ctx context.Context, iCtx pluginsCore.SetupContext, entry k8s.PluginEntry) (*PluginManager, error) {
	if iCtx.EnqueueOwner() == nil {
		return nil, errors.Errorf(errors.PluginInitializationFailed, "Failed to initialize plugin, enqueue Owner cannot be nil or empty.")
	}

	if iCtx.KubeClient() == nil {
		return nil, errors.Errorf(errors.PluginInitializationFailed, "Failed to initialize K8sResource Plugin, Kubeclient cannot be nil!")
	}

	logger.Infof(ctx, "Initializing K8s plugin [%s]", entry.ID)
	src := source.Kind{
		Type: entry.ResourceToWatch,
	}

	ownerKind := iCtx.OwnerKind()
	workflowParentPredicate := func(o metav1.Object) bool {
		ownerReference := metav1.GetControllerOf(o)
		if ownerReference != nil {
			if ownerReference.Kind == ownerKind {
				return true
			}
		}
		return false
	}

	if err := src.InjectCache(iCtx.KubeClient().GetCache()); err != nil {
		logger.Errorf(ctx, "failed to set informers for ObjectType %s", src.String())
		return nil, err
	}

	metricsScope := iCtx.MetricsScope().NewSubScope(entry.ID)
	updateCount := labeled.NewCounter("informer_update", "Update events from informer", metricsScope)
	droppedUpdateCount := labeled.NewCounter("informer_update_dropped", "Update events from informer that have the same resource version", metricsScope)
	genericCount := labeled.NewCounter("informer_generic", "Generic events from informer", metricsScope)

	enqueueOwner := iCtx.EnqueueOwner()
	err := src.Start(
		// Handlers
		handler.Funcs{
			CreateFunc: func(evt event.CreateEvent, q2 workqueue.RateLimitingInterface) {
				logger.Debugf(context.Background(), "Create received for %s, ignoring.", evt.Meta.GetName())
			},
			UpdateFunc: func(evt event.UpdateEvent, q2 workqueue.RateLimitingInterface) {
				if evt.MetaNew == nil {
					logger.Warn(context.Background(), "Received an Update event with nil MetaNew.")
				} else if evt.MetaOld == nil || evt.MetaOld.GetResourceVersion() != evt.MetaNew.GetResourceVersion() {
					newCtx := contextutils.WithNamespace(context.Background(), evt.MetaNew.GetNamespace())
					logger.Debugf(ctx, "Enqueueing owner for updated object [%v/%v]", evt.MetaNew.GetNamespace(), evt.MetaNew.GetName())
					if err := enqueueOwner(k8stypes.NamespacedName{Name: evt.MetaNew.GetName(), Namespace: evt.MetaNew.GetNamespace()}); err != nil {
						logger.Warnf(context.Background(), "Failed to handle Update event for object [%v]", evt.MetaNew.GetName())
					}
					updateCount.Inc(newCtx)
				} else {
					newCtx := contextutils.WithNamespace(context.Background(), evt.MetaNew.GetNamespace())
					droppedUpdateCount.Inc(newCtx)
				}
			},
			DeleteFunc: func(evt event.DeleteEvent, q2 workqueue.RateLimitingInterface) {
				logger.Debugf(context.Background(), "Delete received for %s, ignoring.", evt.Meta.GetName())
			},
			GenericFunc: func(evt event.GenericEvent, q2 workqueue.RateLimitingInterface) {
				logger.Debugf(context.Background(), "Generic received for %s, ignoring.", evt.Meta.GetName())
				genericCount.Inc(ctx)
			},
		},
		// Queue
		// TODO: a more unique workqueue name
		workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(),
			entry.ResourceToWatch.GetObjectKind().GroupVersionKind().Kind),
		// Predicates
		predicate.Funcs{
			CreateFunc: func(createEvent event.CreateEvent) bool {
				return false
			},
			UpdateFunc: func(updateEvent event.UpdateEvent) bool {
				// TODO we should filter out events in case there are no updates observed between the old and new?
				return workflowParentPredicate(updateEvent.MetaNew)
			},
			DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
				return false
			},
			GenericFunc: func(genericEvent event.GenericEvent) bool {
				return workflowParentPredicate(genericEvent.Meta)
			},
		})

	if err != nil {
		return nil, err
	}

	return &PluginManager{
		id:              entry.ID,
		plugin:          entry.Plugin,
		resourceToWatch: entry.ResourceToWatch,
		metrics:         newPluginMetrics(metricsScope),
		kubeClient:      iCtx.KubeClient(),
	}, nil
}
