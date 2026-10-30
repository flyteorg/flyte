package k8s

import (
	"context"
	"fmt"
	"hash/fnv"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	toolscache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	executorv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/executor/pkg/plugin"
	"github.com/flyteorg/flyte/v2/flytestdlib/fastcheck"
	k8sutil "github.com/flyteorg/flyte/v2/flytestdlib/k8s"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/namespaceutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
)

// ActionUpdate represents an update to a TaskAction
type ActionUpdate struct {
	ActionID         *common.ActionIdentifier
	ParentActionName string
	StateJSON        string
	Phase            common.ActionPhase
	OutputUri        string
	IsDeleted        bool
	TaskType         string
	ShortName        string
	ErrorState       *executorv1.ErrorState
	// SignalValue is the resolved signal payload of a condition action, carried
	// inline (never an outputs.pb). Nil for tasks and unsignalled conditions.
	SignalValue *core.Literal
	// Address is where the running task can be reached (UpdateRequest.address).
	// Nil when it reported none, and once the task is terminal.
	Address *common.ActionAddress
}

const (
	labelTerminalStatusRecorded = "flyte.org/terminal-status-recorded"
)

// ActionsClient handles all etcd/K8s TaskAction CR operations for the Actions service.
type ActionsClient struct {
	k8sClient   client.WithWatch
	sharedCache ctrlcache.Cache
	// namespace holds unmapped runs, and runs started before namespaceTemplate was set.
	namespace                 string
	namespaceTemplate         string
	requireExistingNamespaces bool
	// knownNamespaces caches namespaces seen to exist, saving an API call per enqueue.
	knownNamespaces sync.Map
	bufferSize      int
	runClient       workflowconnect.InternalRunServiceClient
	// recordedFilter deduplicates RecordAction calls across watch reconnects.
	recordedFilter fastcheck.Filter

	// Watch management
	mu sync.RWMutex
	// Map parent action name to subscriber channels.
	// Multiple callers may watch the same parent action concurrently.
	// TODO: add a prometheus counter for dropped updates when metrics are wired up
	subscribers map[string]map[chan *ActionUpdate]struct{}
	stopCh      chan struct{}
	watching    bool

	// Worker pool: numWorkers goroutines each own one channel.
	// Events are sharded by the TaskAction name, so per-resource ordering is preserved.
	numWorkers int
	workerChs  []chan watch.Event

	// dispatchedActions tracks TaskAction names with an event already dispatched to a
	// worker channel and not yet processed. Further events for the same action are
	// dropped (coalesced) — the worker reads the latest state from the cache when it
	// runs.
	dispatchedActions   map[string]struct{}
	dispatchedActionsMu sync.Mutex

	// getLatest reads the current TaskAction from the shared cache. Overridable in tests.
	getLatest func(ctx context.Context, key client.ObjectKey) (*executorv1.TaskAction, bool, error)
}

// NewActionsClient creates a new Kubernetes-based actions client.
func NewActionsClient(k8sClient client.WithWatch, sharedCache ctrlcache.Cache, namespace string, bufferSize int, numWorkers int, runClient workflowconnect.InternalRunServiceClient, recordFilterSize int, scope promutils.Scope, namespaceTemplate string, createNamespaces bool) (*ActionsClient, error) {
	if numWorkers <= 0 {
		numWorkers = 1
	}
	// The dedup filter is mandatory, so its size must be a positive value.
	if recordFilterSize <= 0 {
		return nil, fmt.Errorf("actions: recordFilterSize must be positive, got %d", recordFilterSize)
	}
	if scope == nil {
		return nil, fmt.Errorf("actions: metrics scope is required")
	}
	c := &ActionsClient{
		k8sClient:                 k8sClient,
		sharedCache:               sharedCache,
		namespace:                 namespace,
		namespaceTemplate:         namespaceTemplate,
		requireExistingNamespaces: !createNamespaces,
		bufferSize:                bufferSize,
		numWorkers:                numWorkers,
		runClient:                 runClient,
		subscribers:               make(map[string]map[chan *ActionUpdate]struct{}),
	}

	// The dedup filter is mandatory: notifyRunService records on every event type
	// and relies on the filter to keep RecordAction idempotent.
	filter, err := fastcheck.NewOppoBloomFilter(recordFilterSize, scope.NewSubScope("actions_filter"))
	if err != nil {
		return nil, fmt.Errorf("actions: failed to create RecordAction dedup filter (size=%d): %w", recordFilterSize, err)
	}
	c.recordedFilter = filter

	return c, nil
}

// Enqueue creates a TaskAction CR in etcd (via the K8s API).
func (c *ActionsClient) Enqueue(ctx context.Context, action *actions.Action, runSpec *task.RunSpec) error {
	actionID := action.ActionId
	logger.Infof(ctx, "Enqueuing action: %s/%s/%s/%s",
		actionID.Run.Project, actionID.Run.Domain,
		actionID.Run.Name, actionID.Name)

	isRoot := action.ParentActionName == nil || *action.ParentActionName == ""

	switch action.GetSpec().(type) {
	case *actions.Action_Task:
		taskAction := c.newTaskActionCR(actionID, executorv1.ActionTypeTask, isRoot)
		// Set OwnerReference to parent so K8s cascades deletion to children.
		if !isRoot {
			parentTaskAction, err := c.setParentOwnership(ctx, taskAction, actionID.Run, *action.ParentActionName)
			if err != nil {
				return err
			}
			// For child actions, inherit parent's run context
			inheritRunContextFromParentTaskAction(taskAction, parentTaskAction)
		} else {
			if err := c.placeRootAction(ctx, taskAction, actionID); err != nil {
				return err
			}
			// For root action, apply the RunSpec to TaskAction
			applyRunSpecToTaskAction(taskAction, runSpec)
		}

		// Build and set the ActionSpec for the executor.
		actionSpec := buildActionSpec(action, runSpec)
		if err := taskAction.Spec.SetActionSpec(actionSpec); err != nil {
			return fmt.Errorf("failed to set action spec: %w", err)
		}
		taskAction.Spec.CacheKey = extractTaskCacheKey(action)

		// Embed the inline TaskTemplate if present.
		if err := embedTaskTemplate(action, taskAction, runSpec); err != nil {
			return fmt.Errorf("failed to embed task template: %w", err)
		}

		if err := c.k8sClient.Create(ctx, taskAction); err != nil {
			// Surface AlreadyExists so callers can treat a re-enqueue (e.g. from
			// a restarted parent) as already launched.
			if apierrors.IsAlreadyExists(err) {
				return connect.NewError(connect.CodeAlreadyExists,
					fmt.Errorf("TaskAction CR %s already exists: %w", taskAction.Name, err))
			}
			if apierrors.IsNotFound(err) {
				return c.rejectMissingNamespace(taskAction, err)
			}
			return fmt.Errorf("failed to create TaskAction CR %s: %w", taskAction.Name, err)
		}

		logger.Infof(ctx, "Created TaskAction CR: %s", taskAction.Name)
	case *actions.Action_Condition:
		cond := action.GetCondition()
		// Validate the declared type at enqueue time so bad specs fail fast
		// instead of becoming unsignalable CRs.
		if err := validateConditionDeclaredType(cond); err != nil {
			return connect.NewError(connect.CodeInvalidArgument, err)
		}
		taskAction := c.newTaskActionCR(actionID, executorv1.ActionTypeCondition, isRoot)
		if !isRoot {
			if _, err := c.setParentOwnership(ctx, taskAction, actionID.Run, *action.ParentActionName); err != nil {
				return err
			}
		} else if err := c.placeRootAction(ctx, taskAction, actionID); err != nil {
			return err
		}

		actionSpec := buildActionSpec(action, runSpec)
		if err := taskAction.Spec.SetActionSpec(actionSpec); err != nil {
			return fmt.Errorf("failed to set action spec: %w", err)
		}
		taskAction.Spec.ActionType = executorv1.ActionTypeCondition
		condBytes, err := proto.Marshal(cond)
		if err != nil {
			return fmt.Errorf("failed to marshal condition spec: %w", err)
		}
		taskAction.Spec.ConditionSpec = condBytes
		// Unlike task actions, a condition runs no plugin or pod — the CR only
		// carries the ConditionSpec and stays pending until an external signal
		// resolves it. So no TaskTemplate or CacheKey is needed here.

		if err := c.k8sClient.Create(ctx, taskAction); err != nil {
			if apierrors.IsAlreadyExists(err) {
				return connect.NewError(connect.CodeAlreadyExists,
					fmt.Errorf("condition TaskAction CR %s already exists: %w", taskAction.Name, err))
			}
			if apierrors.IsNotFound(err) {
				return c.rejectMissingNamespace(taskAction, err)
			}
			return fmt.Errorf("failed to create condition TaskAction CR %s: %w", taskAction.Name, err)
		}

		logger.Infof(ctx, "Created condition TaskAction CR: %s", taskAction.Name)
	case *actions.Action_Trace:
		// For trace action, we only need to record result in RunService as it is already computed in SDK runtime
		recordReq := &workflow.RecordActionRequest{
			ActionId: actionID,
			Parent:   action.GetParentActionName(),
			InputUri: action.GetInputUri(),
			Group:    action.GetGroup(),
			Subject:  action.GetSubject(),
			Spec: &workflow.RecordActionRequest_Trace{
				Trace: action.GetTrace(),
			},
		}
		_, err := c.runClient.RecordAction(ctx, connect.NewRequest(recordReq))
		if err != nil {
			return err
		}

		// Notify SDK runtime informer
		c.notifySubscribers(ctx, &ActionUpdate{
			ActionID:         actionID,
			ParentActionName: action.GetParentActionName(),
			Phase:            common.ActionPhase_ACTION_PHASE_SUCCEEDED,
			OutputUri:        action.GetTrace().GetOutputs().GetOutputUri(),
			TaskType:         "trace",
		})
	}

	return nil
}

// AbortAction deletes a TaskAction CR from etcd.
// K8s cascades the deletion to all descendants via OwnerReferences.
func (c *ActionsClient) AbortAction(ctx context.Context, actionID *common.ActionIdentifier, reason *string) error {
	taskActionName := buildTaskActionName(actionID)
	logger.Infof(ctx, "Aborting action %s (reason: %v)", taskActionName, reason)

	taskAction := &executorv1.TaskAction{}
	if err := c.getTaskAction(ctx, actionID, taskAction); err != nil {
		return fmt.Errorf("failed to get TaskAction %s: %w", taskActionName, err)
	}

	if err := c.k8sClient.Delete(ctx, taskAction); err != nil {
		return fmt.Errorf("failed to delete TaskAction %s: %w", taskActionName, err)
	}

	logger.Infof(ctx, "Deleted TaskAction %s (descendants will be cascade deleted by K8s)", taskActionName)
	return nil
}

// PutStatus updates the latest attempt metadata for a TaskAction, and where the
// running task can be reached (UpdateRequest.address). An address replaces the
// current one, a nil address keeps it within the same attempt, an empty one
// withdraws it, and a new attempt clears it. The address is recorded with the
// reporting attempt and the TaskAction's system-retry count, and is delivered
// only while both are current (AddressFromStatus). A report from an attempt
// older than the current one changes neither the attempts nor the address.
func (c *ActionsClient) PutStatus(ctx context.Context, actionID *common.ActionIdentifier, attempt uint32, status *workflow.ActionStatus, address *common.ActionAddress) error {
	taskActionName := buildTaskActionName(actionID)

	// Get current TaskAction
	taskAction := &executorv1.TaskAction{}
	if err := c.getTaskAction(ctx, actionID, taskAction); err != nil {
		return fmt.Errorf("failed to get TaskAction %s: %w", taskActionName, err)
	}

	if status == nil {
		return nil
	}

	attempts := status.GetAttempts()
	next := taskAction.Status.Address
	if stale := attempt != 0 && attempt < taskAction.Status.Attempts; stale {
		// A late report from an older attempt: keep the current attempt and address.
		attempts = taskAction.Status.Attempts
	} else {
		next = nextAddress(taskAction.Status.Address, taskAction.Status.Attempts, attempts, taskAction.Status.SystemRetries, address)
	}

	// Skip update if the attempts, cache status and address do not change
	if taskAction.Status.Attempts == attempts &&
		taskAction.Status.CacheStatus == status.GetCacheStatus() &&
		equality.Semantic.DeepEqual(taskAction.Status.Address, next) {
		return nil
	}

	taskAction.Status.Attempts = attempts
	taskAction.Status.CacheStatus = status.GetCacheStatus()
	taskAction.Status.Address = next

	// Update status subresource
	if err := c.k8sClient.Status().Update(ctx, taskAction); err != nil {
		return fmt.Errorf("failed to update TaskAction status %s: %w", taskActionName, err)
	}

	logger.Infof(ctx, "Updated status for TaskAction: %s", taskActionName)
	return nil
}

// Signal delivers the resolved value to a paused condition action by writing
// status.signalValue/signalledBy/signalledAt. The reconciler is the single
// writer of status.conditions[] and flips the CR to Succeeded when it observes
// the value, so validation here is synchronous but the transition is not.
func (c *ActionsClient) Signal(ctx context.Context, actionID *common.ActionIdentifier, value *core.Literal, signalledBy string) error {
	kind, ok := literalPrimitiveKind(value)
	if !ok {
		return connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("signal value must be a concrete bool/int/float/str literal"))
	}
	valueBytes, err := proto.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal signal value: %w", err)
	}
	taskActionName := buildTaskActionName(actionID)

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		taskAction := &executorv1.TaskAction{}
		if err := c.getTaskAction(ctx, actionID, taskAction); err != nil {
			if apierrors.IsNotFound(err) {
				return connect.NewError(connect.CodeNotFound, fmt.Errorf("condition action %s not found", taskActionName))
			}
			return fmt.Errorf("failed to get TaskAction %s: %w", taskActionName, err)
		}
		if taskAction.Spec.ActionType != executorv1.ActionTypeCondition {
			return connect.NewError(connect.CodeFailedPrecondition,
				fmt.Errorf("action %s is not a condition", taskActionName))
		}
		if existing := SignalValueFromStatus(ctx, taskAction); existing != nil {
			if proto.Equal(existing, value) {
				return nil // idempotent retry of the same signal
			}
			return connect.NewError(connect.CodeFailedPrecondition,
				fmt.Errorf("condition %s already signalled with a different value", taskActionName))
		}
		if isTerminalPhase(GetPhaseFromConditions(taskAction)) {
			return connect.NewError(connect.CodeFailedPrecondition,
				fmt.Errorf("condition %s already completed", taskActionName))
		}

		condSpec := &workflow.ConditionAction{}
		if err := proto.Unmarshal(taskAction.Spec.ConditionSpec, condSpec); err != nil {
			return fmt.Errorf("failed to unmarshal condition spec for %s: %w", taskActionName, err)
		}
		if declared := condSpec.GetType().GetSimple(); declared != kind {
			return connect.NewError(connect.CodeInvalidArgument,
				fmt.Errorf("signal value type %s does not match declared condition type %s", kind, declared))
		}

		now := metav1.Now()
		taskAction.Status.SignalValue = valueBytes
		taskAction.Status.SignalledBy = signalledBy
		taskAction.Status.SignalledAt = &now
		if err := c.k8sClient.Status().Update(ctx, taskAction); err != nil {
			return err
		}
		logger.Infof(ctx, "Signalled condition TaskAction %s (by %q)", taskActionName, signalledBy)
		return nil
	})
}

// literalPrimitiveKind maps a concrete signal Literal to the SimpleType it
// would satisfy. ok is false for empty literals and non-primitive values.
func literalPrimitiveKind(value *core.Literal) (core.SimpleType, bool) {
	switch value.GetScalar().GetPrimitive().GetValue().(type) {
	case *core.Primitive_Boolean:
		return core.SimpleType_BOOLEAN, true
	case *core.Primitive_Integer:
		return core.SimpleType_INTEGER, true
	case *core.Primitive_FloatValue:
		return core.SimpleType_FLOAT, true
	case *core.Primitive_StringValue:
		return core.SimpleType_STRING, true
	default:
		return core.SimpleType_NONE, false
	}
}

// ListRunActions lists all TaskActions belonging to a run.
func (c *ActionsClient) ListRunActions(ctx context.Context, runID *common.RunIdentifier) ([]*executorv1.TaskAction, error) {
	items, err := c.listRunTaskActions(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to list TaskActions for run: %w", err)
	}

	result := make([]*executorv1.TaskAction, len(items))
	for i := range items {
		result[i] = &items[i]
	}
	return result, nil
}

// ListChildActions lists all TaskActions that are children of the given parent action
func (c *ActionsClient) ListChildActions(ctx context.Context, parentActionID *common.ActionIdentifier) ([]*executorv1.TaskAction, error) {
	// List all TaskActions in the same run
	items, err := c.listRunTaskActions(ctx, parentActionID.Run)
	if err != nil {
		return nil, fmt.Errorf("failed to list TaskActions: %w", err)
	}

	// Filter for the parent and its children
	var result []*executorv1.TaskAction
	for i := range items {
		action := &items[i]
		// Include the parent action itself
		if action.Spec.ActionName == parentActionID.Name {
			result = append(result, action)
			continue
		}
		// Include direct children
		if action.Spec.ParentActionName != nil && *action.Spec.ParentActionName == parentActionID.Name {
			result = append(result, action)
		}
	}

	return result, nil
}

// GetTaskAction retrieves a specific TaskAction
func (c *ActionsClient) GetTaskAction(ctx context.Context, actionID *common.ActionIdentifier) (*executorv1.TaskAction, error) {
	taskActionName := buildTaskActionName(actionID)

	taskAction := &executorv1.TaskAction{}
	if err := c.getTaskAction(ctx, actionID, taskAction); err != nil {
		return nil, fmt.Errorf("failed to get TaskAction %s: %w", taskActionName, err)
	}

	return taskAction, nil
}

// Subscribe creates a new subscription channel for action updates scoped to the given (run, parent action).
// subscriberKey scopes a subscription to a single (run, parent action) pair.
// The parent action name alone is NOT unique across runs -- every run's root
// action is named "a0" -- so keying on it alone broadcasts each run's child
// updates to every other run's watch stream. Same-named children then collide
// in the SDK informer cache and clobber each other's phase, wedging the parent.
// Run names are only unique within a project and domain, so those are part of the key too.
func subscriberKey(runID *common.RunIdentifier, parentActionName string) string {
	return runID.GetProject() + "/" + runID.GetDomain() + "/" + runID.GetName() + "/" + parentActionName
}

func (c *ActionsClient) Subscribe(runID *common.RunIdentifier, parentActionName string) chan *ActionUpdate {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := subscriberKey(runID, parentActionName)
	ch := make(chan *ActionUpdate, c.bufferSize)
	if c.subscribers[key] == nil {
		c.subscribers[key] = make(map[chan *ActionUpdate]struct{})
	}
	c.subscribers[key][ch] = struct{}{}
	return ch
}

// Unsubscribe removes the given channel from the subscription list for the (run, parent action)
func (c *ActionsClient) Unsubscribe(runID *common.RunIdentifier, parentActionName string, ch chan *ActionUpdate) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := subscriberKey(runID, parentActionName)
	if channels, ok := c.subscribers[key]; ok {
		delete(channels, ch)
		close(ch)
		if len(channels) == 0 {
			delete(c.subscribers, key)
		}
	}
}

// StartWatching starts watching TaskAction resources and notifies all subscribers.
// It requires a shared controller-runtime cache.
func (c *ActionsClient) StartWatching(ctx context.Context) error {
	c.mu.Lock()
	if c.watching {
		c.mu.Unlock()
		return nil
	}
	c.watching = true
	c.stopCh = make(chan struct{})
	stopCh := c.stopCh // capture before releasing so workers reference the right channel
	c.workerChs = make([]chan watch.Event, c.numWorkers)
	c.dispatchedActions = make(map[string]struct{})
	if c.getLatest == nil {
		c.getLatest = c.getLatestFromCache
	}
	for i := range c.workerChs {
		c.workerChs[i] = make(chan watch.Event, c.bufferSize)
		go c.worker(ctx, c.workerChs[i], stopCh)
	}
	c.mu.Unlock()

	logger.Infof(ctx, "Starting TaskAction watcher (workers: %d)", c.numWorkers)

	if c.sharedCache == nil {
		return fmt.Errorf("shared cache is required for TaskAction informer")
	}

	return c.setupInformer(ctx)
}

func (c *ActionsClient) setupInformer(ctx context.Context) error {
	informer, err := c.sharedCache.GetInformer(ctx, &executorv1.TaskAction{})
	if err != nil {
		return fmt.Errorf("failed to get TaskAction informer: %w", err)
	}

	_, err = informer.AddEventHandler(toolscache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			taskAction, ok := obj.(*executorv1.TaskAction)
			if !ok || c.shouldSkipTaskAction(taskAction) {
				return
			}
			c.dispatchEvent(taskAction, watch.Added)
		},
		UpdateFunc: func(_, newObj interface{}) {
			taskAction, ok := newObj.(*executorv1.TaskAction)
			if !ok || c.shouldSkipTaskAction(taskAction) {
				return
			}
			c.dispatchEvent(taskAction, watch.Modified)
		},
		DeleteFunc: func(obj interface{}) {
			// The informer may deliver a DeletedFinalStateUnknown tombstone
			// when a delete event was missed; unwrap it first.
			if tombstone, ok := obj.(toolscache.DeletedFinalStateUnknown); ok {
				obj = tombstone.Obj
			}
			taskAction, ok := obj.(*executorv1.TaskAction)
			if !ok || c.shouldSkipTaskAction(taskAction) {
				return
			}
			c.dispatchEvent(taskAction, watch.Deleted)
		},
	})
	if err != nil {
		return fmt.Errorf("failed to add TaskAction informer handler: %w", err)
	}

	return nil
}

// worker drains its event channel and handles each event inline.
// Each worker owns a disjoint shard of TaskAction names, so per-resource
// event ordering is preserved across the pool.
func (c *ActionsClient) worker(ctx context.Context, ch <-chan watch.Event, stopCh <-chan struct{}) {
	for {
		select {
		case <-stopCh:
			return
		case <-ctx.Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			c.handleWatchEvent(ctx, event)
		}
	}
}

// dispatchEvent routes an informer event to the worker responsible for the
// TaskAction, using FNV-32a hashing for consistent sharding. TaskActions are keyed
// by namespace and name, since the same name can exist in two namespaces.
func (c *ActionsClient) dispatchEvent(taskAction *executorv1.TaskAction, eventType watch.EventType) {
	if taskAction == nil {
		return
	}

	key := client.ObjectKeyFromObject(taskAction).String()
	h := fnv.New32a()
	_, _ = h.Write([]byte(key)) // FNV Write never returns an error
	shard := h.Sum32() % uint32(c.numWorkers)

	if eventType == watch.Deleted {
		c.workerChs[shard] <- watch.Event{Type: eventType, Object: taskAction.DeepCopy()}
		return
	}

	c.dispatchedActionsMu.Lock()
	if _, already := c.dispatchedActions[key]; already {
		c.dispatchedActionsMu.Unlock()
		return
	}
	c.dispatchedActions[key] = struct{}{}
	c.dispatchedActionsMu.Unlock()

	c.workerChs[shard] <- watch.Event{Type: eventType, Object: taskAction.DeepCopy()}
}

func (c *ActionsClient) handleWatchEvent(ctx context.Context, event watch.Event) {
	taskAction, ok := event.Object.(*executorv1.TaskAction)
	if !ok {
		logger.Warnf(ctx, "received non-TaskAction object in watch event: %T", event.Object)
		return
	}

	key := client.ObjectKeyFromObject(taskAction)

	// Deletes carry the tombstone (the object is gone from the cache); process it directly.
	if event.Type == watch.Deleted {
		c.dispatchedActionsMu.Lock()
		delete(c.dispatchedActions, key.String())
		c.dispatchedActionsMu.Unlock()
		c.handleTaskActionEvent(ctx, taskAction, watch.Deleted)
		return
	}

	c.dispatchedActionsMu.Lock()
	delete(c.dispatchedActions, key.String())
	c.dispatchedActionsMu.Unlock()

	latest, found, err := c.getLatest(ctx, key)
	if err != nil {
		logger.Warnf(ctx, "coalesced watch: failed to read latest TaskAction %s: %v", taskAction.Name, err)
		return
	}
	if !found {
		return // object already gone; a delete event (if any) handles the terminal state
	}
	if c.shouldSkipTaskAction(latest) {
		return
	}
	c.handleTaskActionEvent(ctx, latest, event.Type)
}

// getLatestFromCache reads the current TaskAction from the shared informer cache.
func (c *ActionsClient) getLatestFromCache(ctx context.Context, key client.ObjectKey) (*executorv1.TaskAction, bool, error) {
	ta := &executorv1.TaskAction{}
	if err := c.sharedCache.Get(ctx, key, ta); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return ta, true, nil
}

func (c *ActionsClient) handleTaskActionEvent(ctx context.Context, taskAction *executorv1.TaskAction, eventType watch.EventType) {
	update := buildActionUpdate(ctx, taskAction, eventType)
	if update == nil {
		return
	}

	c.notifySubscribers(ctx, update)
	c.notifyRunService(ctx, taskAction, update, eventType)
}

func buildActionUpdate(ctx context.Context, taskAction *executorv1.TaskAction, eventType watch.EventType) *ActionUpdate {
	var parentName string
	if taskAction.Spec.ParentActionName != nil {
		parentName = *taskAction.Spec.ParentActionName
	}

	// Determine a short name: use spec.ShortName if set, otherwise extract from template ID
	shortName := taskAction.Spec.ShortName
	if shortName == "" && len(taskAction.Spec.TaskTemplate) > 0 {
		shortName = extractShortNameFromTemplate(taskAction.Spec.TaskTemplate)
	}

	phase := GetPhaseFromConditions(taskAction)
	if eventType == watch.Deleted && !isTerminalPhase(phase) {
		// Only force ABORTED if the action wasn't already in a terminal phase.
		// Otherwise a missed-delete tombstone or post-terminal CR cleanup would
		// overwrite a recorded Succeeded/Failed status with Aborted.
		phase = common.ActionPhase_ACTION_PHASE_ABORTED
	}

	return &ActionUpdate{
		ActionID: &common.ActionIdentifier{
			Run: &common.RunIdentifier{
				Project: taskAction.Spec.Project,
				Domain:  taskAction.Spec.Domain,
				Name:    taskAction.Spec.RunName,
			},
			Name: taskAction.Spec.ActionName,
		},
		ParentActionName: parentName,
		StateJSON:        taskAction.Status.StateJSON,
		Phase:            phase,
		OutputUri:        BuildOutputUri(ctx, taskAction),
		IsDeleted:        eventType == watch.Deleted,
		TaskType:         taskAction.Spec.TaskType,
		ShortName:        shortName,
		ErrorState:       taskAction.Status.ErrorState,
		SignalValue:      SignalValueFromStatus(ctx, taskAction),
		Address:          AddressFromStatus(taskAction),
	}
}

// nextAddress returns the address a task holds after an update for attempt
// `attempts`: a new address replaces the held one (recorded with the attempt and
// the TaskAction's system-retry count), an empty one withdraws it, none keeps it
// within the attempt, and a new attempt clears it.
func nextAddress(held *executorv1.ActionAddress, heldAttempts, attempts, systemRetries uint32, address *common.ActionAddress) *executorv1.ActionAddress {
	if address != nil {
		if isEmptyAddress(address) {
			return nil
		}
		return addressToStatus(address, attempts, systemRetries)
	}
	if attempts != heldAttempts {
		return nil
	}
	return held
}

// isEmptyAddress reports an address with nothing set: a withdrawal.
func isEmptyAddress(a *common.ActionAddress) bool {
	p := a.GetPod()
	return a.GetPath() == "" && a.GetUrl() == "" && a.GetCluster() == "" &&
		(p == nil || (p.GetNamespace() == "" && p.GetName() == "" && p.GetIp() == "" && p.GetPort() == 0))
}

func addressToStatus(a *common.ActionAddress, attempt, systemRetries uint32) *executorv1.ActionAddress {
	out := &executorv1.ActionAddress{
		Path: a.GetPath(), URL: a.GetUrl(), Cluster: a.GetCluster(),
		Attempt: attempt, SystemRetries: systemRetries,
	}
	if p := a.GetPod(); p != nil {
		out.Pod = &executorv1.PodAddress{Namespace: p.GetNamespace(), Name: p.GetName(), IP: p.GetIp(), Port: p.GetPort()}
	}
	return out
}

// AddressFromStatus is where the running task can be reached, or nil when it
// reported none, is terminal, or has moved on from the attempt or pod that
// reported it (a retry, or a system-retry relaunch, the controller started).
func AddressFromStatus(taskAction *executorv1.TaskAction) *common.ActionAddress {
	a := taskAction.Status.Address
	if a == nil || isTerminalPhase(GetPhaseFromConditions(taskAction)) {
		return nil
	}
	if a.Attempt != taskAction.Status.Attempts || a.SystemRetries != taskAction.Status.SystemRetries {
		return nil
	}
	out := &common.ActionAddress{Path: a.Path, Url: a.URL, Cluster: a.Cluster}
	if a.Pod != nil {
		out.Pod = &common.PodAddress{Namespace: a.Pod.Namespace, Name: a.Pod.Name, Ip: a.Pod.IP, Port: a.Pod.Port}
	}
	return out
}

// SignalValueFromStatus unmarshals the condition signal value persisted on the
// CR status. Returns nil when absent (tasks, unsignalled conditions) or invalid.
func SignalValueFromStatus(ctx context.Context, taskAction *executorv1.TaskAction) *core.Literal {
	if len(taskAction.Status.SignalValue) == 0 {
		return nil
	}
	value := &core.Literal{}
	if err := proto.Unmarshal(taskAction.Status.SignalValue, value); err != nil {
		logger.Warnf(ctx, "failed to unmarshal signal value for %s: %v", taskAction.Name, err)
		return nil
	}
	return value
}

func (c *ActionsClient) shouldSkipTaskAction(taskAction *executorv1.TaskAction) bool {
	if taskAction == nil {
		return true
	}
	return taskAction.GetLabels()[labelTerminalStatusRecorded] == "true"
}

// notifySubscribers sends an update to all subscribers
func (c *ActionsClient) notifySubscribers(ctx context.Context, update *ActionUpdate) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := subscriberKey(update.ActionID.GetRun(), update.ParentActionName)
	for ch := range c.subscribers[key] {
		select {
		case ch <- update:
		default:
			logger.Warnf(ctx, "subscriber channel full, dropping update for parent action: %s", update.ParentActionName)
		}
	}
}

// notifyRunService forwards a watch event to the internal run service.
// On ADDED events it calls RecordAction to create the DB record.
// On all events it calls UpdateActionStatus (when phase is meaningful) to update the actions table.
func (c *ActionsClient) notifyRunService(ctx context.Context, taskAction *executorv1.TaskAction, update *ActionUpdate, eventType watch.EventType) {
	if c.runClient == nil {
		return
	}
	actionKey := recordedFilterKey(taskAction)
	isDuplicate := c.recordedFilter.Contains(ctx, actionKey)
	if isDuplicate {
		logger.Debugf(ctx, "Skipping duplicate RecordAction for %s", update.ActionID.Name)
	} else {
		recordReq := &workflow.RecordActionRequest{
			ActionId: update.ActionID,
			Parent:   update.ParentActionName,
			InputUri: taskAction.Spec.InputURI,
			Group:    taskAction.Spec.Group,
		}
		if taskAction.Spec.ActionType == executorv1.ActionTypeCondition {
			condSpec := &workflow.ConditionAction{}
			if err := proto.Unmarshal(taskAction.Spec.ConditionSpec, condSpec); err != nil {
				logger.Warnf(ctx, "Failed to unmarshal condition spec for %s: %v", update.ActionID.Name, err)
			} else {
				recordReq.Spec = &workflow.RecordActionRequest_Condition{Condition: condSpec}
			}
		} else if taskAction.Spec.TaskType != "" {
			ta := &workflow.TaskAction{
				Id: &task.TaskIdentifier{
					Project: taskAction.Spec.Project,
					Domain:  taskAction.Spec.Domain,
				},
			}
			// Deserialize TaskTemplate to build TaskSpec
			if len(taskAction.Spec.TaskTemplate) > 0 {
				var tmpl core.TaskTemplate
				if err := proto.Unmarshal(taskAction.Spec.TaskTemplate, &tmpl); err == nil {
					if tmplID := tmpl.GetId(); tmplID != nil {
						ta.Id.Name = tmplID.GetName()
						ta.Id.Version = tmplID.GetVersion()
					}
					ta.Spec = &task.TaskSpec{
						TaskTemplate: &tmpl,
						ShortName:    taskAction.Spec.ShortName,
					}
				}
			}
			recordReq.Spec = &workflow.RecordActionRequest_Task{
				Task: ta,
			}
		}
		if _, err := c.runClient.RecordAction(ctx, connect.NewRequest(recordReq)); err != nil {
			logger.Warnf(ctx, "Failed to record action in run service for %s: %v", update.ActionID.Name, err)
		} else {
			c.recordedFilter.Add(ctx, actionKey)
		}
	}

	if update.Phase != common.ActionPhase_ACTION_PHASE_UNSPECIFIED {
		status := &workflow.ActionStatus{
			Phase:       update.Phase,
			Attempts:    taskAction.Status.Attempts,
			CacheStatus: taskAction.Status.CacheStatus,
		}
		// Carry the action's true start (its CRD creationTimestamp) so the run service
		// computes duration from it. Otherwise created_at is stamped whenever this event
		// is recorded, which — under coalesced/backlogged events — can collapse to the
		// terminal time and report a near-zero duration for a long-held action.
		if !taskAction.CreationTimestamp.IsZero() {
			status.StartTime = timestamppb.New(taskAction.CreationTimestamp.Time)
		}
		statusReq := &workflow.UpdateActionStatusRequest{
			ActionId: update.ActionID,
			Status:   status,
		}
		// On terminal SUCCEEDED of a signalled condition, ship the resolved
		// value and actor to the run-service DB.
		if update.Phase == common.ActionPhase_ACTION_PHASE_SUCCEEDED && update.SignalValue != nil {
			statusReq.Output = update.SignalValue
			if taskAction.Status.SignalledBy != "" {
				statusReq.Principal = &common.EnrichedIdentity{
					Principal: &common.EnrichedIdentity_User{
						User: &common.User{Id: &common.UserIdentifier{Subject: taskAction.Status.SignalledBy}},
					},
				}
			}
		}
		if resp, err := c.runClient.UpdateActionStatus(ctx, connect.NewRequest(statusReq)); err != nil {
			logger.Warnf(ctx, "Failed to update action status in run service for %s: %v", update.ActionID.Name, err)
		} else if code := resp.Msg.GetStatus().GetCode(); code != 0 {
			logger.Warnf(ctx, "Run service rejected action status update for %s: code=%d message=%s",
				update.ActionID.Name, code, resp.Msg.GetStatus().GetMessage())
		} else if isTerminalPhase(update.Phase) && !update.IsDeleted {
			// Skip label patching for deleted CRs — the patch would always fail
			// with "not found" since the object is already gone.
			if err := c.markTerminalStatusRecorded(ctx, taskAction); err != nil {
				logger.Warnf(ctx, "Failed to mark terminal status recorded for %s: %v", update.ActionID.Name, err)
			}
		}
	}
}

// StopWatching stops the TaskAction watcher
func (c *ActionsClient) StopWatching() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.watching && c.stopCh != nil {
		close(c.stopCh)
		c.watching = false
	}
}

// GetPhaseFromConditions extracts the phase from TaskAction conditions.
func GetPhaseFromConditions(taskAction *executorv1.TaskAction) common.ActionPhase {
	for _, cond := range taskAction.Status.Conditions {
		switch cond.Type {
		case string(executorv1.ConditionTypeSucceeded):
			if cond.Status == "True" {
				return common.ActionPhase_ACTION_PHASE_SUCCEEDED
			}
		case string(executorv1.ConditionTypeFailed):
			if cond.Status == "True" {
				if cond.Reason == string(executorv1.ConditionReasonTimedOut) {
					return common.ActionPhase_ACTION_PHASE_TIMED_OUT
				}
				return common.ActionPhase_ACTION_PHASE_FAILED
			}
		case string(executorv1.ConditionTypeProgressing):
			if cond.Status == "True" {
				switch cond.Reason {
				case string(executorv1.ConditionReasonQueued):
					return common.ActionPhase_ACTION_PHASE_QUEUED
				case string(executorv1.ConditionReasonInitializing):
					return common.ActionPhase_ACTION_PHASE_INITIALIZING
				case string(executorv1.ConditionReasonExecuting):
					return common.ActionPhase_ACTION_PHASE_RUNNING
				case string(executorv1.ConditionReasonPaused):
					return common.ActionPhase_ACTION_PHASE_PAUSED
				}
			}
		}
	}
	return common.ActionPhase_ACTION_PHASE_UNSPECIFIED
}

func isTerminalPhase(phase common.ActionPhase) bool {
	return phase == common.ActionPhase_ACTION_PHASE_SUCCEEDED ||
		phase == common.ActionPhase_ACTION_PHASE_FAILED ||
		phase == common.ActionPhase_ACTION_PHASE_ABORTED ||
		phase == common.ActionPhase_ACTION_PHASE_TIMED_OUT ||
		phase == common.ActionPhase_ACTION_PHASE_RECOVERED
}

func (c *ActionsClient) markTerminalStatusRecorded(ctx context.Context, taskAction *executorv1.TaskAction) error {
	if c.k8sClient == nil || taskAction == nil || c.shouldSkipTaskAction(taskAction) {
		return nil
	}

	original := taskAction.DeepCopy()
	patched := taskAction.DeepCopy()
	labels := patched.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[labelTerminalStatusRecorded] = "true"
	patched.SetLabels(labels)

	if err := c.k8sClient.Patch(ctx, patched, client.MergeFrom(original)); err != nil {
		return err
	}
	return nil
}

// newTaskActionCR builds the CR skeleton (name, labels) shared by all action
// types. Its namespace is set by placeRootAction or setParentOwnership.
func (c *ActionsClient) newTaskActionCR(actionID *common.ActionIdentifier, actionType string, isRoot bool) *executorv1.TaskAction {
	return &executorv1.TaskAction{
		ObjectMeta: metav1.ObjectMeta{
			Name: buildTaskActionName(actionID),
			Labels: map[string]string{
				"flyte.org/project":     actionID.Run.Project,
				"flyte.org/domain":      actionID.Run.Domain,
				"flyte.org/run":         actionID.Run.Name,
				"flyte.org/action":      actionID.Name,
				"flyte.org/action-type": actionType,
				"flyte.org/is-root":     fmt.Sprintf("%t", isRoot),
			},
		},
		Spec: executorv1.TaskActionSpec{},
	}
}

// setParentOwnership resolves the parent TaskAction and wires an OwnerReference
// on the child so K8s cascades deletion. The child is created in its parent's
// namespace, since an OwnerReference cannot cross namespaces. Returns the parent CR.
func (c *ActionsClient) setParentOwnership(ctx context.Context, child *executorv1.TaskAction, runID *common.RunIdentifier, parentActionName string) (*executorv1.TaskAction, error) {
	parentID := &common.ActionIdentifier{
		Run:  runID,
		Name: parentActionName,
	}
	parentName := buildTaskActionName(parentID)

	parent := &executorv1.TaskAction{}
	if err := c.getTaskAction(ctx, parentID, parent); err != nil {
		return nil, fmt.Errorf("failed to get parent TaskAction %s: %w", parentName, err)
	}

	child.Namespace = parent.Namespace
	blockOwnerDeletion := true
	child.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion:         "flyte.org/v1",
			Kind:               "TaskAction",
			Name:               parent.Name,
			UID:                parent.UID,
			BlockOwnerDeletion: &blockOwnerDeletion,
		},
	}
	return parent, nil
}

// recordedFilterKey includes the namespace, since the same run and action names can exist in two.
func recordedFilterKey(taskAction *executorv1.TaskAction) []byte {
	return []byte(client.ObjectKeyFromObject(taskAction).String())
}

func (c *ActionsClient) placeRootAction(ctx context.Context, taskAction *executorv1.TaskAction, actionID *common.ActionIdentifier) error {
	runID := actionID.GetRun()
	namespace, err := namespaceutils.MappedNamespace(c.namespaceTemplate, c.namespace, runID.GetOrg(), runID.GetProject(), runID.GetDomain())
	if err != nil {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}
	// A run is resubmitted by name (the scheduler fires a run again), so a run
	// started before the template was set may already have its root in the
	// default namespace. Report it as launched rather than start a second root.
	if namespace != c.namespace {
		err := c.getLegacyTaskAction(ctx, actionID, &executorv1.TaskAction{})
		if err == nil {
			return connect.NewError(connect.CodeAlreadyExists,
				fmt.Errorf("TaskAction CR %s already exists in namespace %s", taskAction.Name, c.namespace))
		}
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to check for TaskAction CR %s in namespace %s: %w", taskAction.Name, c.namespace, err)
		}
	}
	if err := c.ensureNamespace(ctx, namespace); err != nil {
		return err
	}
	taskAction.Namespace = namespace
	return nil
}

func (c *ActionsClient) ensureNamespace(ctx context.Context, namespace string) error {
	if _, ok := c.knownNamespaces.Load(namespace); ok {
		return nil
	}
	if c.requireExistingNamespaces {
		err := c.k8sClient.Get(ctx, client.ObjectKey{Name: namespace}, &corev1.Namespace{})
		if apierrors.IsNotFound(err) {
			return connect.NewError(connect.CodeFailedPrecondition,
				fmt.Errorf("namespace %q does not exist; create it before running in this project and domain", namespace))
		}
		if err != nil {
			return fmt.Errorf("failed to check namespace %s: %w", namespace, err)
		}
	} else if err := k8sutil.EnsureNamespaceExists(ctx, c.k8sClient, namespace); err != nil {
		return fmt.Errorf("failed to ensure namespace %s: %w", namespace, err)
	}
	c.knownNamespaces.Store(namespace, struct{}{})
	return nil
}

// rejectMissingNamespace forgets a namespace deleted after it was cached, and returns
// FailedPrecondition so CreateRun fails the run instead of leaving it queued.
func (c *ActionsClient) rejectMissingNamespace(taskAction *executorv1.TaskAction, err error) error {
	c.knownNamespaces.Delete(taskAction.Namespace)
	return connect.NewError(connect.CodeFailedPrecondition,
		fmt.Errorf("failed to create TaskAction CR %s: %w", taskAction.Name, err))
}

// lookupNamespaces also returns the client's namespace, which holds runs started before
// the template was set.
func (c *ActionsClient) lookupNamespaces(runID *common.RunIdentifier) []string {
	namespace, err := namespaceutils.MappedNamespace(c.namespaceTemplate, c.namespace, runID.GetOrg(), runID.GetProject(), runID.GetDomain())
	if err != nil || namespace == c.namespace {
		return []string{c.namespace}
	}
	return []string{namespace, c.namespace}
}

func (c *ActionsClient) getTaskAction(ctx context.Context, actionID *common.ActionIdentifier, taskAction *executorv1.TaskAction) error {
	namespaces := c.lookupNamespaces(actionID.GetRun())
	err := c.k8sClient.Get(ctx, client.ObjectKey{Name: buildTaskActionName(actionID), Namespace: namespaces[0]}, taskAction)
	if len(namespaces) == 1 || !apierrors.IsNotFound(err) {
		return err
	}
	return c.getLegacyTaskAction(ctx, actionID, taskAction)
}

// getLegacyTaskAction reads from the client's namespace. All projects share it and a
// TaskAction name holds only the run name, so another project's or domain's is NotFound.
func (c *ActionsClient) getLegacyTaskAction(ctx context.Context, actionID *common.ActionIdentifier, taskAction *executorv1.TaskAction) error {
	name := buildTaskActionName(actionID)
	if err := c.k8sClient.Get(ctx, client.ObjectKey{Name: name, Namespace: c.namespace}, taskAction); err != nil {
		return err
	}
	labels := taskAction.GetLabels()
	if labels["flyte.org/project"] != actionID.GetRun().GetProject() || labels["flyte.org/domain"] != actionID.GetRun().GetDomain() {
		return apierrors.NewNotFound(executorv1.GroupVersion.WithResource("taskactions").GroupResource(), name)
	}
	return nil
}

func (c *ActionsClient) listRunTaskActions(ctx context.Context, runID *common.RunIdentifier) ([]executorv1.TaskAction, error) {
	var items []executorv1.TaskAction
	for _, namespace := range c.lookupNamespaces(runID) {
		taskActionList := &executorv1.TaskActionList{}
		if err := c.k8sClient.List(ctx, taskActionList,
			client.InNamespace(namespace),
			client.MatchingLabels{
				"flyte.org/project": runID.Project,
				"flyte.org/domain":  runID.Domain,
				"flyte.org/run":     runID.Name,
			},
		); err != nil {
			return nil, err
		}
		items = append(items, taskActionList.Items...)
	}
	return items, nil
}

// validateConditionDeclaredType enforces that a condition's declared type is a
// signalable simple primitive (bool/int/float/string), mirroring cloud's
// enqueue-time validation.
func validateConditionDeclaredType(cond *workflow.ConditionAction) error {
	switch cond.GetType().GetSimple() {
	case core.SimpleType_BOOLEAN, core.SimpleType_INTEGER, core.SimpleType_FLOAT, core.SimpleType_STRING:
		return nil
	default:
		return fmt.Errorf("condition %q declares unsupported type %s; must be one of bool/int/float/str", cond.GetName(), cond.GetType())
	}
}

// buildTaskActionName generates a Kubernetes-compliant name for the TaskAction.
// For root actions (where action name == run name), the name is <run-id>-a0.
// For child actions, the name is <run-id>-<action-id>.
func buildTaskActionName(actionID *common.ActionIdentifier) string {
	isRoot := actionID.Name == actionID.Run.Name
	if isRoot {
		return fmt.Sprintf("%s-a0", actionID.Run.Name)
	}
	return fmt.Sprintf("%s-%s", actionID.Run.Name, actionID.Name)
}

// BuildOutputUri computes the action-specific output URI from the TaskAction spec.
// It uses the same path structure as the executor's ComputeActionOutputPath so that
// the SDK can find outputs written by the executor.
func BuildOutputUri(ctx context.Context, ta *executorv1.TaskAction) string {
	if ta.Spec.RunOutputBase == "" {
		return ""
	}
	attempt := ta.Status.Attempts
	if attempt == 0 { // if attempts is not set, default to 1
		attempt = 1
	}
	prefix, err := plugin.ComputeActionOutputPath(ctx, ta.Namespace, ta.Name, ta.Spec.RunOutputBase, ta.Spec.ActionName, attempt)
	if err != nil {
		return ""
	}
	return string(prefix)
}

// InitScheme adds the executor API types to the scheme
func InitScheme() error {
	return executorv1.AddToScheme(scheme.Scheme)
}

// buildActionSpec converts an actions.Action into the workflow.ActionSpec expected by the executor.
func buildActionSpec(action *actions.Action, runSpec *task.RunSpec) *workflow.ActionSpec {
	actionSpec := &workflow.ActionSpec{
		ActionId:         action.ActionId,
		ParentActionName: action.ParentActionName,
		RunSpec:          runSpec,
		InputUri:         action.InputUri,
		RunOutputBase:    action.RunOutputBase,
		Group:            action.Group,
	}

	switch spec := action.Spec.(type) {
	case *actions.Action_Task:
		actionSpec.Spec = &workflow.ActionSpec_Task{Task: spec.Task}
	case *actions.Action_Trace:
		actionSpec.Spec = &workflow.ActionSpec_Trace{Trace: spec.Trace}
	case *actions.Action_Condition:
		actionSpec.Spec = &workflow.ActionSpec_Condition{Condition: spec.Condition}
	}

	return actionSpec
}

func applyRunSpecToTaskAction(taskAction *executorv1.TaskAction, runSpec *task.RunSpec) {
	if runSpec == nil {
		taskAction.Spec.EnvVars = nil
		taskAction.Spec.Interruptible = nil
		taskAction.Spec.PodTemplateName = ""
		return
	}

	taskAction.Spec.EnvVars = keyValuePairsToMap(runSpec.GetEnvs().GetValues())
	taskAction.Spec.PodTemplateName = runSpec.GetDefaultSettings().GetPodTemplateName()
	if runSpec.GetInterruptible() != nil {
		value := runSpec.GetInterruptible().GetValue()
		taskAction.Spec.Interruptible = &value
	} else {
		taskAction.Spec.Interruptible = nil
	}

	for key, value := range runSpec.GetLabels().GetValues() {
		if _, exists := taskAction.Labels[key]; !exists {
			taskAction.Labels[key] = value
		}
	}

	if len(runSpec.GetAnnotations().GetValues()) > 0 {
		if taskAction.Annotations == nil {
			taskAction.Annotations = make(map[string]string, len(runSpec.GetAnnotations().GetValues()))
		}
		for key, value := range runSpec.GetAnnotations().GetValues() {
			taskAction.Annotations[key] = value
		}
	}
}

func inheritRunContextFromParentTaskAction(taskAction *executorv1.TaskAction, parentTaskAction *executorv1.TaskAction) {
	if taskAction == nil || parentTaskAction == nil {
		return
	}
	taskAction.Spec.EnvVars = cloneStringMap(parentTaskAction.Spec.EnvVars)
	taskAction.Spec.PodTemplateName = parentTaskAction.Spec.PodTemplateName
	if len(parentTaskAction.Annotations) > 0 {
		if taskAction.Annotations == nil {
			taskAction.Annotations = map[string]string{}
		}
		for k, v := range cloneStringMap(parentTaskAction.Annotations) {
			if _, exists := taskAction.Annotations[k]; !exists {
				taskAction.Annotations[k] = v
			}
		}
	}
	if len(parentTaskAction.Labels) > 0 {
		if taskAction.Labels == nil {
			taskAction.Labels = map[string]string{}
		}
		for k, v := range cloneStringMap(parentTaskAction.Labels) {
			if _, exists := taskAction.Labels[k]; !exists {
				taskAction.Labels[k] = v
			}
		}
	}
	if parentTaskAction.Spec.Interruptible != nil {
		v := *parentTaskAction.Spec.Interruptible
		taskAction.Spec.Interruptible = &v
	} else {
		taskAction.Spec.Interruptible = nil
	}
}

func cloneStringMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]string, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func keyValuePairsToMap(values []*core.KeyValuePair) map[string]string {
	if len(values) == 0 {
		return nil
	}

	cloned := make(map[string]string, len(values))
	for _, kv := range values {
		if kv == nil {
			continue
		}
		cloned[kv.GetKey()] = kv.GetValue()
	}
	return cloned
}

func extractTaskCacheKey(action *actions.Action) string {
	taskSpec, ok := action.Spec.(*actions.Action_Task)
	if !ok || taskSpec.Task == nil {
		return ""
	}

	if taskSpec.Task.CacheKey == nil {
		return ""
	}

	return taskSpec.Task.CacheKey.Value
}

// runStartTimeTemplateVar is the placeholder the SDK (>= 2.3.6) emits in a task's container args
// for the run start time. The backend substitutes it at TaskAction creation so the value surfaces
// as flyte.ctx().run_start_time. See flyteorg/flyte-sdk#1100. Unlike --run-name / --name (which the
// SDK reads from the RUN_NAME / ACTION_NAME env vars when left unsubstituted), --run-start-time has
// no env-var fallback, so the value must be baked into the args here.
const runStartTimeTemplateVar = "{{.runStartTime}}"

// embedTaskTemplate serializes the inline TaskTemplate from the Action into the CR spec,
// substituting the {{.runStartTime}} container-arg placeholder with runSpec.run_start_time when
// set. Templates produced by older SDKs omit the placeholder, so the substitution is a no-op.
func embedTaskTemplate(action *actions.Action, taskAction *executorv1.TaskAction, runSpec *task.RunSpec) error {
	taskSpec, ok := action.Spec.(*actions.Action_Task)
	if !ok || taskSpec.Task == nil || taskSpec.Task.Spec == nil || taskSpec.Task.Spec.TaskTemplate == nil {
		// Non-task actions do not carry an inline template.
		return nil
	}

	tmpl := taskSpec.Task.Spec.TaskTemplate
	taskAction.Spec.TaskType = tmpl.Type
	taskAction.Spec.ShortName = taskSpec.Task.Spec.ShortName

	tmpl = substituteRunStartTime(tmpl, runSpec.GetRunStartTime())
	tmpl = applyPodTemplateName(tmpl, taskAction.Spec.PodTemplateName)

	data, err := proto.Marshal(tmpl)
	if err != nil {
		return fmt.Errorf("failed to marshal task template: %w", err)
	}
	taskAction.Spec.TaskTemplate = data
	return nil
}

// substituteRunStartTime returns a TaskTemplate whose container args have the {{.runStartTime}}
// placeholder replaced with ts formatted as RFC3339 UTC. It returns the input unchanged when ts is
// nil or the template has no container args. The template is cloned before mutation so the caller's
// proto — which may be persisted separately as the registered task spec — is not affected.
func substituteRunStartTime(tmpl *core.TaskTemplate, ts *timestamppb.Timestamp) *core.TaskTemplate {
	if ts == nil {
		return tmpl
	}
	if tmpl.GetContainer() == nil || len(tmpl.GetContainer().GetArgs()) == 0 {
		return tmpl
	}
	value := ts.AsTime().UTC().Format(time.RFC3339)
	cloned := proto.Clone(tmpl).(*core.TaskTemplate)
	args := cloned.GetContainer().GetArgs()
	for i, arg := range args {
		args[i] = strings.ReplaceAll(arg, runStartTimeTemplateVar, value)
	}
	return cloned
}

// applyPodTemplateName returns a TaskTemplate whose metadata names the given pod template.
// It returns the input unchanged when the name is empty or the task already named one itself,
// so an explicit choice always wins. The template is cloned before mutation so the caller's
// proto - which may be persisted separately as the registered task spec - is not affected.
func applyPodTemplateName(tmpl *core.TaskTemplate, name string) *core.TaskTemplate {
	if name == "" {
		return tmpl
	}
	if tmpl.GetMetadata().GetPodTemplateName() != "" {
		return tmpl
	}

	cloned := proto.Clone(tmpl).(*core.TaskTemplate)
	if cloned.Metadata == nil {
		cloned.Metadata = &core.TaskMetadata{}
	}
	cloned.Metadata.PodTemplateName = name
	return cloned
}

// extractShortNameFromTemplate extracts a human-readable function name from a serialized TaskTemplate.
// It splits on '.' and returns the last part.
func extractShortNameFromTemplate(templateBytes []byte) string {
	tmpl := &core.TaskTemplate{}
	if err := proto.Unmarshal(templateBytes, tmpl); err != nil {
		return ""
	}
	if tmpl.GetId() == nil {
		return ""
	}
	name := tmpl.GetId().GetName()
	if name == "" {
		return ""
	}
	// Split on '.' and take the last part
	parts := strings.Split(name, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return name
}
