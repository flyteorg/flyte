package flytek8s

import (
	"context"
	"time"

	"github.com/lyft/flytestdlib/contextutils"

	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/promutils/labeled"

	"github.com/lyft/flyteplugins/go/tasks/v1/events"

	k8stypes "k8s.io/apimachinery/pkg/types"

	"strings"

	eventErrors "github.com/lyft/flyteidl/clients/go/events/errors"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteplugins/go/tasks/v1/errors"
	"github.com/lyft/flyteplugins/go/tasks/v1/types"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/storage"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// A generic task executor for k8s-resource reliant tasks.
type K8sTaskExecutor struct {
	types.OutputsResolver
	recorder     types.EventRecorder
	id           string
	handler      K8sResourceHandler
	resyncPeriod time.Duration
	// Supplied on Initialization
	// TODO decide the right place to put these interfaces or late-bind them?
	store           storage.ComposedProtobufStore
	resourceToWatch runtime.Object
	metrics         K8sTaskExecutorMetrics
}

type K8sTaskExecutorMetrics struct {
	Scope           promutils.Scope
	GetCacheMiss    labeled.StopWatch
	GetCacheHit     labeled.StopWatch
	GetAPILatency   labeled.StopWatch
	ResourceDeleted labeled.Counter
}

type ownerRegisteringHandler struct {
	ownerKind    string
	enqueueOwner types.EnqueueOwner
}

// A common handle for all k8s-resource reliant task executors that push workflow id on the work queue.
func (h ownerRegisteringHandler) Handle(ctx context.Context, k8sObject runtime.Object) error {
	object := k8sObject.(metav1.Object)
	ownerReference := metav1.GetControllerOf(object)
	namespace := object.GetNamespace()
	if ownerReference == nil {
		return nil
	}

	if ownerReference.Kind == h.ownerKind {
		return h.enqueueOwner(k8stypes.NamespacedName{Name: ownerReference.Name, Namespace: namespace})
	}

	// Had a log line here but it was way too verbose.  Every pod owned by something other than a
	// Flyte workflow would come up.
	return nil
}

func (e *K8sTaskExecutor) GetID() types.TaskExecutorName {
	return e.id
}

func (e *K8sTaskExecutor) GetProperties() types.ExecutorProperties {
	return types.ExecutorProperties{}
}

func (e *K8sTaskExecutor) Initialize(ctx context.Context, params types.ExecutorInitializationParameters) error {

	if params.DataStore == nil {
		return errors.Errorf(errors.PluginInitializationFailed, "Failed to initialize plugin, data store cannot be nil or empty.")
	}

	if params.EventRecorder == nil {
		return errors.Errorf(errors.PluginInitializationFailed, "Failed to initialize plugin, event recorder cannot be nil or empty.")
	}

	if params.EnqueueOwner == nil {
		return errors.Errorf(errors.PluginInitializationFailed, "Failed to initialize plugin, enqueue Owner cannot be nil or empty.")
	}

	e.store = params.DataStore
	e.recorder = params.EventRecorder
	e.OutputsResolver = types.NewOutputsResolver(params.DataStore)

	metricScope := params.MetricsScope.NewSubScope(e.GetID())
	e.metrics = K8sTaskExecutorMetrics{
		Scope: metricScope,
		GetCacheMiss: labeled.NewStopWatch("get_cache_miss", "Cache miss on get resource calls.",
			time.Millisecond, metricScope),
		GetCacheHit: labeled.NewStopWatch("get_cache_hit", "Cache miss on get resource calls.",
			time.Millisecond, metricScope),
		GetAPILatency: labeled.NewStopWatch("get_api", "Latency for APIServer Get calls.",
			time.Millisecond, metricScope),
		ResourceDeleted: labeled.NewCounter("pods_deleted", "Counts how many times CheckTaskStatus is"+
			" called with a deleted resource.", metricScope),
	}

	return RegisterResource(ctx, e.resourceToWatch, ownerRegisteringHandler{
		enqueueOwner: params.EnqueueOwner,
		ownerKind:    params.OwnerKind,
	})
}

func (e K8sTaskExecutor) HandleTaskSuccess(ctx context.Context, taskCtx types.TaskContext) (types.TaskStatus, error) {
	errorPath := taskCtx.GetErrorFile()
	metadata, err := e.store.Head(ctx, errorPath)
	if err != nil {
		return types.TaskStatusRetryableFailure(errors.Wrapf(errors.MetadataAccessFailed, err,
			"failed to read error file")), nil
	}

	if metadata.Exists() {
		if metadata.Size() > MaxMetadataPayloadSizeBytes {
			return types.TaskStatusPermanentFailure(errors.Errorf(errors.MetadataTooLarge,
				"metadata file is too large [%v] bytes, max allowed [%v] bytes", metadata.Size(),
				MaxMetadataPayloadSizeBytes)), nil
		}
		errorDoc := &core.ErrorDocument{}
		err = e.store.ReadProtobuf(ctx, errorPath, errorDoc)
		if err != nil {
			if storage.IsNotFound(err) {
				return types.TaskStatusRetryableFailure(errors.Wrapf(errors.MetadataAccessFailed, err,
					"metadata file not found but head returned exists")), nil
			}

			return types.TaskStatusUndefined, errors.Wrapf(errors.DownstreamSystemError, err,
				"failed to read error data from task @[%s]", errorPath)
		}

		if errorDoc.Error == nil {
			return types.TaskStatusRetryableFailure(errors.Errorf(errors.MetadataAccessFailed,
				"error not formatted correctly, missing error @path [%v]", errorPath)), nil
		}

		if errorDoc.Error.Kind == core.ContainerError_RECOVERABLE {
			return types.TaskStatusRetryableFailure(errors.Errorf(errorDoc.Error.Code,
				"user-error. Message: [%s]", errorDoc.Error.Message)), nil
		}

		return types.TaskStatusPermanentFailure(errors.Errorf(errorDoc.Error.Code,
			"user-error. Message: [%s]", errorDoc.Error.Message)), nil
	}

	return types.TaskStatusSucceeded, nil
}

func (e *K8sTaskExecutor) StartTask(ctx context.Context, taskCtx types.TaskContext, task *core.TaskTemplate, inputs *core.LiteralMap) (
	types.TaskStatus, error) {

	o, err := e.handler.BuildResource(ctx, taskCtx, task, inputs)
	if err != nil {
		return types.TaskStatusUndefined, err
	}

	AddObjectMetadata(taskCtx, o)
	logger.Infof(ctx, "Creating Object: Type:[%v], Object:[%v/%v]", o.GroupVersionKind(), o.GetNamespace(), o.GetName())

	err = instance.kubeClient.Create(ctx, o)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		if k8serrors.IsBadRequest(err) {
			logger.Errorf(ctx, "Bad Request. [%+v]", o)
			return types.TaskStatusUndefined, err
		}
		if k8serrors.IsForbidden(err) {
			if strings.Contains(err.Error(), "exceeded quota") {
				// TODO: Quota errors are retried forever, it would be good to have support for backoff strategy.
				logger.Warnf(ctx, "Failed to launch job, resource quota exceeded. Err: %v", err)
				return types.TaskStatusNotReadyFailure(err), nil
			}
			return types.TaskStatusPermanentFailure(err), nil
		}
		logger.Errorf(ctx, "Failed to launch job, system error. Err: %v", err)
		return types.TaskStatusUndefined, err
	}
	status := types.TaskStatusQueued

	ev := events.CreateEvent(taskCtx, status, nil)
	err = e.recorder.RecordTaskEvent(ctx, ev)
	if err != nil && eventErrors.IsEventAlreadyInTerminalStateError(err) {
		return types.TaskStatusPermanentFailure(errors.Wrapf(errors.TaskEventRecordingFailed, err,
			"failed to record task event. phase mis-match between Propeller %v and Control Plane.", &status.Phase)), nil
	} else if err != nil {
		return types.TaskStatusUndefined, errors.Wrapf(errors.TaskEventRecordingFailed, err,
			"failed to record start task event")
	}

	return status, nil
}

func (e *K8sTaskExecutor) getResource(ctx context.Context, taskCtx types.TaskContext, o K8sResource) (types.TaskStatus,
	*events.TaskEventInfo, error) {

	nsName := k8stypes.NamespacedName{Namespace: o.GetNamespace(), Name: o.GetName()}
	start := time.Now()
	// Attempt to get resource from informer cache, if not found, retrieve it from API server.
	err := instance.informersCache.Get(ctx, nsName, o)
	if err != nil && IsK8sObjectNotExists(err) {
		e.metrics.GetCacheMiss.Observe(ctx, start, time.Now())
		e.metrics.GetAPILatency.Time(ctx, func() {
			err = instance.kubeClient.Get(ctx, nsName, o)
		})
	} else if err == nil {
		e.metrics.GetCacheHit.Observe(ctx, start, time.Now())
	}

	if err != nil {
		if IsK8sObjectNotExists(err) {
			// This happens sometimes because a node gets removed and K8s deletes the pod. This will result in a
			// Pod does not exist error. This should be retried using the retry policy
			logger.Warningf(ctx, "Failed to find the Resource with name: %v. Error: %v", nsName, err)
			return types.TaskStatusRetryableFailure(err), nil, nil
		}

		logger.Warningf(ctx, "Failed to retrieve Resource Details with name: %v. Error: %v", nsName, err)
		return types.TaskStatusUndefined, nil, err
	}

	return e.handler.GetTaskStatus(ctx, taskCtx, o)
}

func (e *K8sTaskExecutor) CheckTaskStatus(ctx context.Context, taskCtx types.TaskContext, task *core.TaskTemplate) (
	types.TaskStatus, error) {

	o, err := e.handler.BuildIdentityResource(ctx, taskCtx)
	finalStatus := types.TaskStatus{
		Phase:        taskCtx.GetPhase(),
		PhaseVersion: taskCtx.GetPhaseVersion(),
	}

	var info *events.TaskEventInfo

	if err != nil {
		logger.Warningf(ctx, "Failed to build the Resource with name: %v. Error: %v",
			taskCtx.GetTaskExecutionID().GetGeneratedName(), err)
		finalStatus = types.TaskStatusPermanentFailure(err)
	} else {
		AddObjectMetadata(taskCtx, o)
		finalStatus, info, err = e.getResource(ctx, taskCtx, o)
		if err != nil {
			return types.TaskStatusUndefined, err
		}

		if o.GetDeletionTimestamp() != nil {
			e.metrics.ResourceDeleted.Inc(ctx)
		}

		if finalStatus.Phase == types.TaskPhaseSucceeded {
			finalStatus, err = e.HandleTaskSuccess(ctx, taskCtx)
			if err != nil {
				return types.TaskStatusUndefined, err
			}
		}
	}

	if finalStatus.Phase != taskCtx.GetPhase() {
		ev := events.CreateEvent(taskCtx, finalStatus, info)
		err := e.recorder.RecordTaskEvent(ctx, ev)
		if err != nil && eventErrors.IsEventAlreadyInTerminalStateError(err) {
			return types.TaskStatusPermanentFailure(errors.Wrapf(errors.TaskEventRecordingFailed, err,
				"failed to record task event. phase mis-match between Propeller %v and Control Plane.", &ev.Phase)), nil
		} else if err != nil {
			return types.TaskStatusUndefined, errors.Wrapf(errors.TaskEventRecordingFailed, err,
				"failed to record state transition [%v] -> [%v]", taskCtx.GetPhase(), finalStatus.Phase)
		}
	}

	// This must happen after sending admin event. It's safe against partial failures because if the event failed, we will
	// simply retry in the next round. If the event succeeded but this failed, we will try again the next round to send
	// the same event (idempotent) and then come here again...
	if finalStatus.Phase.IsTerminal() && len(o.GetFinalizers()) > 0 {
		err = e.ClearFinalizers(ctx, o)
		if err != nil {
			return types.TaskStatusUndefined, err
		}
	}

	// If the object has been deleted, that is, it has a deletion timestamp, but is not in a terminal state, we should
	// mark the task as a retryable failure.  We've seen this happen when a kubelet disappears - all pods running on
	// the node are marked with a deletionTimestamp, but our finalizers prevent the pod from being deleted.
	// This can also happen when a user deletes a Pod directly.
	if !finalStatus.Phase.IsTerminal() && o.GetDeletionTimestamp() != nil && len(o.GetFinalizers()) > 0 {
		err = e.ClearFinalizers(ctx, o)
		if err != nil {
			return types.TaskStatusUndefined, err
		}
		return types.TaskStatusRetryableFailure(finalStatus.Err), nil
	}

	return finalStatus, nil
}

func (e *K8sTaskExecutor) KillTask(ctx context.Context, taskCtx types.TaskContext, reason string) error {
	logger.Infof(ctx, "KillTask invoked for %v, clearing finalizers.", taskCtx.GetTaskExecutionID().GetGeneratedName())

	o, err := e.handler.BuildIdentityResource(ctx, taskCtx)
	if err != nil {
		logger.Warningf(ctx, "Failed to build the Resource with name: %v. Error: %v", taskCtx.GetTaskExecutionID().GetGeneratedName(), err)
		return err
	}

	AddObjectMetadata(taskCtx, o)

	// Retrieve the object from cache/etcd to get the last known version.
	_, _, err = e.getResource(ctx, taskCtx, o)
	if err != nil {
		return err
	}

	// Clear finalizers
	err = e.ClearFinalizers(ctx, o)
	if err != nil {
		return err
	}

	if e.handler.GetProperties().DeleteResourceOnAbort {
		return instance.kubeClient.Delete(ctx, o)
	}

	return nil
}

func (e *K8sTaskExecutor) ClearFinalizers(ctx context.Context, o K8sResource) error {
	if len(o.GetFinalizers()) > 0 {
		o.SetFinalizers([]string{})
		err := instance.kubeClient.Update(ctx, o)

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

// Creates a K8s generic task executor. This provides an easier way to build task executors that create K8s resources.
func NewK8sTaskExecutorForResource(id string, resourceToWatch runtime.Object, handler K8sResourceHandler,
	resyncPeriod time.Duration) *K8sTaskExecutor {

	return &K8sTaskExecutor{
		id:              id,
		handler:         handler,
		resourceToWatch: resourceToWatch,
		resyncPeriod:    resyncPeriod,
	}
}

func init() {
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)
}
