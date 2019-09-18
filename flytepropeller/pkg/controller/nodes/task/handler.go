package task

import (
	"context"
	"fmt"
	"reflect"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/lyft/flytepropeller/pkg/controller/executors"

	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"

	"github.com/lyft/flytestdlib/promutils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/lyft/flyteidl/clients/go/events"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	pluginsV1 "github.com/lyft/flyteplugins/go/tasks/v1/types"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/promutils/labeled"
	"github.com/lyft/flytestdlib/storage"
	errors2 "github.com/pkg/errors"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/controller/catalog"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/errors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
	"github.com/lyft/flytepropeller/pkg/utils"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const IDMaxLength = 50

// TODO handle retries
type taskContext struct {
	taskExecutionID    taskExecutionID
	dataDir            storage.DataReference
	workflow           v1alpha1.WorkflowMeta
	node               v1alpha1.ExecutableNode
	status             v1alpha1.ExecutableTaskNodeStatus
	serviceAccountName string
}

func (t *taskContext) GetCustomState() pluginsV1.CustomState {
	return t.status.GetCustomState()
}

func (t *taskContext) GetPhase() pluginsV1.TaskPhase {
	return t.status.GetPhase()
}

func (t *taskContext) GetPhaseVersion() uint32 {
	return t.status.GetPhaseVersion()
}

type taskExecutionID struct {
	execName string
	id       core.TaskExecutionIdentifier
}

func (te taskExecutionID) GetID() core.TaskExecutionIdentifier {
	return te.id
}

func (te taskExecutionID) GetGeneratedName() string {
	return te.execName
}

func (t *taskContext) GetOwnerID() types.NamespacedName {
	return t.workflow.GetK8sWorkflowID()
}

func (t *taskContext) GetTaskExecutionID() pluginsV1.TaskExecutionID {
	return t.taskExecutionID
}

func (t *taskContext) GetDataDir() storage.DataReference {
	return t.dataDir
}

func (t *taskContext) GetInputsFile() storage.DataReference {
	return v1alpha1.GetInputsFile(t.dataDir)
}

func (t *taskContext) GetOutputsFile() storage.DataReference {
	return v1alpha1.GetOutputsFile(t.dataDir)
}

func (t *taskContext) GetErrorFile() storage.DataReference {
	return v1alpha1.GetOutputErrorFile(t.dataDir)
}

func (t *taskContext) GetNamespace() string {
	return t.workflow.GetNamespace()
}

func (t *taskContext) GetOwnerReference() v1.OwnerReference {
	return t.workflow.NewControllerRef()
}

func (t *taskContext) GetOverrides() pluginsV1.TaskOverrides {
	return t.node
}

func (t *taskContext) GetLabels() map[string]string {
	return t.workflow.GetLabels()
}

func (t *taskContext) GetAnnotations() map[string]string {
	return t.workflow.GetAnnotations()
}

func (t *taskContext) GetK8sServiceAccount() string {
	return t.serviceAccountName
}

type metrics struct {
	pluginPanics             labeled.Counter
	unsupportedTaskType      labeled.Counter
	discoveryPutFailureCount labeled.Counter
	discoveryGetFailureCount labeled.Counter
	discoveryMissCount       labeled.Counter
	discoveryHitCount        labeled.Counter
	pluginExecutionLatency   labeled.StopWatch

	// TODO We should have a metric to capture custom state size
}

type taskHandler struct {
	taskFactory   Factory
	recorder      events.TaskEventRecorder
	enqueueWf     v1alpha1.EnqueueWorkflow
	store         *storage.DataStore
	scope         promutils.Scope
	catalogClient catalog.Client
	kubeClient    executors.Client
	metrics       *metrics
}

func (h *taskHandler) GetTaskExecutorContext(ctx context.Context, w v1alpha1.ExecutableWorkflow,
	node v1alpha1.ExecutableNode) (pluginsV1.Executor, v1alpha1.ExecutableTask, pluginsV1.TaskContext, error) {

	taskID := node.GetTaskID()
	if taskID == nil {
		return nil, nil, nil, errors.Errorf(errors.BadSpecificationError, node.GetID(), "Task Id not set for NodeKind `Task`")
	}
	task, err := w.GetTask(*taskID)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(errors.BadSpecificationError, node.GetID(), err, "Unable to find task for taskId: [%v]", *taskID)
	}

	exec, err := h.taskFactory.GetTaskExecutor(task.TaskType())
	if err != nil {
		h.metrics.unsupportedTaskType.Inc(ctx)
		return nil, nil, nil, errors.Wrapf(errors.UnsupportedTaskTypeError, node.GetID(), err,
			"Unable to find taskExecutor for taskId: [%v]. TaskType: [%v]", *taskID, task.TaskType())
	}

	nodeStatus := w.GetNodeExecutionStatus(node.GetID())
	id := core.TaskExecutionIdentifier{
		TaskId:       task.CoreTask().Id,
		RetryAttempt: nodeStatus.GetAttempts(),
		NodeExecutionId: &core.NodeExecutionIdentifier{
			NodeId:      node.GetID(),
			ExecutionId: w.GetExecutionID().WorkflowExecutionIdentifier,
		},
	}

	uniqueID, err := utils.FixedLengthUniqueIDForParts(IDMaxLength, w.GetName(), node.GetID(), strconv.Itoa(int(id.RetryAttempt)))
	if err != nil {
		// SHOULD never really happen
		return nil, nil, nil, err
	}

	taskNodeStatus := nodeStatus.GetTaskNodeStatus()
	if taskNodeStatus == nil {
		mutableTaskNodeStatus := nodeStatus.GetOrCreateTaskStatus()
		taskNodeStatus = mutableTaskNodeStatus
	}

	return exec, task, &taskContext{
		taskExecutionID:    taskExecutionID{execName: uniqueID, id: id},
		dataDir:            nodeStatus.GetDataDir(),
		workflow:           w,
		node:               node,
		status:             taskNodeStatus,
		serviceAccountName: w.GetServiceAccountName(),
	}, nil
}

func (h *taskHandler) ExtractOutput(ctx context.Context, w v1alpha1.ExecutableWorkflow, n v1alpha1.ExecutableNode,
	bindToVar handler.VarName) (values *core.Literal, err error) {
	t, task, taskCtx, err := h.GetTaskExecutorContext(ctx, w, n)
	if err != nil {
		return nil, errors.Wrapf(errors.CausedByError, n.GetID(), err, "failed to create TaskCtx")
	}

	l, err := t.ResolveOutputs(ctx, taskCtx, bindToVar)
	if err != nil {
		return nil, errors.Wrapf(errors.CausedByError, n.GetID(), err,
			"failed to resolve output [%v] from task of type [%v]", bindToVar, task.TaskType())
	}

	return l[bindToVar], nil
}

func (h *taskHandler) StartNode(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode, nodeInputs *handler.Data) (handler.Status, error) {
	t, task, taskCtx, err := h.GetTaskExecutorContext(ctx, w, node)
	if err != nil {
		return handler.StatusFailed(errors.Wrapf(errors.CausedByError, node.GetID(), err, "failed to create TaskCtx")), nil
	}

	logger.Infof(ctx, "Executor type: [%v]. Properties: finalizer[%v]. disable[%v].", reflect.TypeOf(t).String(), t.GetProperties().RequiresFinalizer, t.GetProperties().DisableNodeLevelCaching)
	if iface := task.CoreTask().Interface; task.CoreTask().Metadata.Discoverable && iface != nil && iface.Outputs != nil && len(iface.Outputs.Variables) > 0 {
		if t.GetProperties().DisableNodeLevelCaching {
			logger.Infof(ctx, "Executor has Node-Level caching disabled. Skipping.")
		} else if resp, err := h.catalogClient.Get(ctx, task.CoreTask(), taskCtx.GetInputsFile()); err != nil {
			if taskStatus, ok := status.FromError(err); ok && taskStatus.Code() == codes.NotFound {
				h.metrics.discoveryMissCount.Inc(ctx)
				logger.Infof(ctx, "Artifact not found in cache. Executing Task.")
			} else {
				h.metrics.discoveryGetFailureCount.Inc(ctx)
				logger.Errorf(ctx, "Catalog cache check failed. Executing Task. Err: %v", err.Error())
			}
		} else if resp != nil {
			h.metrics.discoveryHitCount.Inc(ctx)

			logger.Debugf(ctx, "Outputs found in Catalog cache %+v", resp)
			if err := h.store.WriteProtobuf(ctx, taskCtx.GetOutputsFile(), storage.Options{}, resp); err != nil {
				logger.Errorf(ctx, "failed to write data to Storage, err: %v", err.Error())
				return handler.StatusUndefined, errors.Wrapf(errors.CausedByError, node.GetID(), err, "failed to copy cached results for task.")
			}

			// SetCached.
			w.GetNodeExecutionStatus(node.GetID()).SetCached()
			return handler.StatusSuccess, nil
		} else {
			// Nil response and Nil error
			h.metrics.discoveryGetFailureCount.Inc(ctx)
			return handler.StatusUndefined, errors.Wrapf(errors.CatalogCallFailed, node.GetID(), err, "Nil catalog response. Failed to check Catalog for previous results")
		}
	}

	var taskStatus pluginsV1.TaskStatus
	func() {
		defer func() {
			if r := recover(); r != nil {
				h.metrics.pluginPanics.Inc(ctx)
				stack := debug.Stack()
				err = fmt.Errorf("panic when executing a plugin for TaskType [%s]. Stack: [%s]", task.TaskType(), string(stack))
				logger.Errorf(ctx, "Panic in plugin for TaskType [%s]", task.TaskType())
			}
		}()
		t := h.metrics.pluginExecutionLatency.Start(ctx)
		taskStatus, err = t.StartTask(ctx, taskCtx, task.CoreTask(), nodeInputs)
		t.Stop()
	}()

	if err != nil {
		return handler.StatusUndefined, errors.Wrapf(errors.CausedByError, node.GetID(), err, "failed to start task [retry attempt: %d]", taskCtx.GetTaskExecutionID().GetID().RetryAttempt)
	}

	nodeStatus := w.GetNodeExecutionStatus(node.GetID())
	taskNodeStatus := nodeStatus.GetOrCreateTaskStatus()
	taskNodeStatus.SetPhase(taskStatus.Phase)
	taskNodeStatus.SetPhaseVersion(taskStatus.PhaseVersion)
	taskNodeStatus.SetCustomState(taskStatus.State)

	logger.Debugf(ctx, "Started Task Node")
	return ConvertTaskPhaseToHandlerStatus(taskStatus)
}

func ConvertTaskPhaseToHandlerStatus(taskStatus pluginsV1.TaskStatus) (handler.Status, error) {
	// TODO handle retryable failure
	switch taskStatus.Phase {
	case pluginsV1.TaskPhaseNotReady:
		return handler.StatusQueued.WithOccurredAt(taskStatus.OccurredAt), nil
	case pluginsV1.TaskPhaseQueued, pluginsV1.TaskPhaseRunning:
		return handler.StatusRunning.WithOccurredAt(taskStatus.OccurredAt), nil
	case pluginsV1.TaskPhasePermanentFailure:
		return handler.StatusFailed(taskStatus.Err).WithOccurredAt(taskStatus.OccurredAt), nil
	case pluginsV1.TaskPhaseRetryableFailure:
		return handler.StatusRetryableFailure(taskStatus.Err).WithOccurredAt(taskStatus.OccurredAt), nil
	case pluginsV1.TaskPhaseSucceeded:
		return handler.StatusSuccess.WithOccurredAt(taskStatus.OccurredAt), nil
	default:
		return handler.StatusUndefined, errors.Errorf(errors.IllegalStateError, "received unknown task phase. [%s]", taskStatus.Phase.String())
	}
}

func (h *taskHandler) CheckNodeStatus(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode, prevNodeStatus v1alpha1.ExecutableNodeStatus) (handler.Status, error) {
	t, task, taskCtx, err := h.GetTaskExecutorContext(ctx, w, node)
	if err != nil {
		return handler.StatusFailed(errors.Wrapf(errors.CausedByError, node.GetID(), err, "Failed to create TaskCtx")), nil
	}

	var taskStatus pluginsV1.TaskStatus
	func() {
		defer func() {
			if r := recover(); r != nil {
				h.metrics.pluginPanics.Inc(ctx)
				stack := debug.Stack()
				err = fmt.Errorf("panic when executing a plugin for TaskType [%s]. Stack: [%s]", task.TaskType(), string(stack))
				logger.Errorf(ctx, "Panic in plugin for TaskType [%s]", task.TaskType())
			}
		}()
		t := h.metrics.pluginExecutionLatency.Start(ctx)
		taskStatus, err = t.CheckTaskStatus(ctx, taskCtx, task.CoreTask())
		t.Stop()
	}()

	if err != nil {
		logger.Warnf(ctx, "Failed to check status")
		return handler.StatusUndefined, errors.Wrapf(errors.CausedByError, node.GetID(), err, "failed to check status")
	}

	nodeStatus := w.GetNodeExecutionStatus(node.GetID())
	taskNodeStatus := nodeStatus.GetOrCreateTaskStatus()
	taskNodeStatus.SetPhase(taskStatus.Phase)
	taskNodeStatus.SetPhaseVersion(taskStatus.PhaseVersion)
	taskNodeStatus.SetCustomState(taskStatus.State)

	return ConvertTaskPhaseToHandlerStatus(taskStatus)
}

func (h *taskHandler) HandleNodeSuccess(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode) (handler.Status, error) {
	t, task, taskCtx, err := h.GetTaskExecutorContext(ctx, w, node)
	if err != nil {
		return handler.StatusFailed(errors.Wrapf(errors.CausedByError, node.GetID(), err, "Failed to create TaskCtx")), nil
	}

	// If the task interface has outputs, validate that the outputs file was written.
	if iface := task.CoreTask().Interface; task.TaskType() != "container_array" && iface != nil && iface.Outputs != nil && len(iface.Outputs.Variables) > 0 {
		if metadata, err := h.store.Head(ctx, taskCtx.GetOutputsFile()); err != nil {
			return handler.StatusUndefined, errors.Wrapf(errors.CausedByError, node.GetID(), err, "failed to HEAD task outputs file.")
		} else if !metadata.Exists() {
			return handler.StatusRetryableFailure(errors.Errorf(errors.OutputsNotFoundError, node.GetID(),
				"Outputs not found for task type %s, looking for output file %s", task.TaskType(), taskCtx.GetOutputsFile())), nil
		}

		// ignores discovery write failures
		if task.CoreTask().Metadata.Discoverable && !t.GetProperties().DisableNodeLevelCaching {
			taskExecutionID := taskCtx.GetTaskExecutionID().GetID()
			if err2 := h.catalogClient.Put(ctx, task.CoreTask(), &taskExecutionID, taskCtx.GetInputsFile(), taskCtx.GetOutputsFile()); err2 != nil {
				h.metrics.discoveryPutFailureCount.Inc(ctx)
				logger.Errorf(ctx, "Failed to write results to catalog. Err: %v", err2)
			} else {
				logger.Debugf(ctx, "Successfully cached results - Task [%s]", task.CoreTask().GetId())
			}
		}
	}
	return handler.StatusSuccess, nil
}

func (h *taskHandler) HandleFailingNode(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode) (handler.Status, error) {
	return handler.StatusFailed(errors.Errorf(errors.IllegalStateError, node.GetID(), "A regular Task node cannot enter a failing state")), nil
}

func (h *taskHandler) Initialize(ctx context.Context) error {
	logger.Infof(ctx, "Initializing taskHandler")
	enqueueFn := func(ownerId types.NamespacedName) error {
		h.enqueueWf(ownerId.String())
		return nil
	}

	initParams := pluginsV1.ExecutorInitializationParameters{
		CatalogClient: h.catalogClient,
		EventRecorder: h.recorder,
		DataStore:     h.store,
		EnqueueOwner:  enqueueFn,
		OwnerKind:     v1alpha1.FlyteWorkflowKind,
		MetricsScope:  h.scope,
	}

	for _, r := range h.taskFactory.ListAllTaskExecutors() {
		logger.Infof(ctx, "Initializing Executor [%v]", r.GetID())
		// Inject a RuntimeClient if the executor needs one.
		if _, err := inject.ClientInto(h.kubeClient.GetClient(), r); err != nil {
			return errors2.Wrapf(err, "Failed to initialize [%v]", r.GetID())
		}

		if _, err := inject.CacheInto(h.kubeClient.GetCache(), r); err != nil {
			return errors2.Wrapf(err, "Failed to initialize [%v]", r.GetID())
		}

		err := r.Initialize(ctx, initParams)
		if err != nil {
			return errors2.Wrapf(err, "Failed to Initialize TaskExecutor [%v]", r.GetID())
		}
	}

	logger.Infof(ctx, "taskHandler Initialization complete")
	return nil
}

func (h *taskHandler) AbortNode(ctx context.Context, w v1alpha1.ExecutableWorkflow, node v1alpha1.ExecutableNode) error {
	t, _, taskCtx, err := h.GetTaskExecutorContext(ctx, w, node)
	if err != nil {
		return errors.Wrapf(errors.CausedByError, node.GetID(), err, "failed to create TaskCtx")
	}

	err = t.KillTask(ctx, taskCtx, "Node aborted")
	if err != nil {
		return errors.Wrapf(errors.CausedByError, node.GetID(), err, "failed to abort task")
	}
	// TODO: Do we need to update the Node status to Failed here as well ?
	logger.Infof(ctx, "Invoked KillTask on Task Node.")
	return nil
}

func NewTaskHandlerForFactory(eventSink events.EventSink, store *storage.DataStore, enqueueWf v1alpha1.EnqueueWorkflow,
	tf Factory, catalogClient catalog.Client, kubeClient executors.Client, scope promutils.Scope) handler.IFace {

	// create a recorder for the plugins
	eventsRecorder := utils.NewPluginTaskEventRecorder(events.NewTaskEventRecorder(eventSink, scope))
	return &taskHandler{
		taskFactory:   tf,
		recorder:      eventsRecorder,
		enqueueWf:     enqueueWf,
		store:         store,
		scope:         scope,
		catalogClient: catalogClient,
		kubeClient:    kubeClient,
		metrics: &metrics{
			pluginPanics:             labeled.NewCounter("plugin_panic", "Task plugin paniced when trying to execute a task.", scope),
			unsupportedTaskType:      labeled.NewCounter("unsupported_tasktype", "No task plugin configured for task type", scope),
			discoveryHitCount:        labeled.NewCounter("discovery_hit_count", "Task cached in Discovery", scope),
			discoveryMissCount:       labeled.NewCounter("discovery_miss_count", "Task not cached in Discovery", scope),
			discoveryPutFailureCount: labeled.NewCounter("discovery_put_failure_count", "Discovery Put failure count", scope),
			discoveryGetFailureCount: labeled.NewCounter("discovery_get_failure_count", "Discovery Get faillure count", scope),
			pluginExecutionLatency:   labeled.NewStopWatch("plugin_exec_latecny", "Time taken to invoke plugin for one round", time.Microsecond, scope),
		},
	}
}

func New(eventSink events.EventSink, store *storage.DataStore, enqueueWf v1alpha1.EnqueueWorkflow, revalPeriod time.Duration,
	catalogClient catalog.Client, kubeClient executors.Client, scope promutils.Scope) handler.IFace {

	return NewTaskHandlerForFactory(eventSink, store, enqueueWf, NewFactory(revalPeriod),
		catalogClient, kubeClient, scope.NewSubScope("task"))
}
