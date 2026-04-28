package k8s

import (
	"context"
	"fmt"
	"hash/fnv"
	"strings"
	"sync"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	toolscache "k8s.io/client-go/tools/cache"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	executorv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/executor/pkg/plugin"
	"github.com/flyteorg/flyte/v2/flytestdlib/fastcheck"
	k8sutil "github.com/flyteorg/flyte/v2/flytestdlib/k8s"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
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
}

const (
	labelTerminalStatusRecorded = "flyte.org/terminal-status-recorded"
	// flyteNamespace is the single Kubernetes namespace used for all Flyte resources.
	flyteNamespace = "flyte"
)

// ActionsClient handles all etcd/K8s TaskAction CR operations for the Actions service.
type ActionsClient struct {
	k8sClient   client.WithWatch
	sharedCache ctrlcache.Cache
	bufferSize  int
	runClient   workflowconnect.InternalRunServiceClient
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
	// Events are sharded by TaskAction name so per-resource ordering is preserved.
	numWorkers int
	workerChs  []chan watch.Event
}

// NewActionsClient creates a new Kubernetes-based actions client.
func NewActionsClient(k8sClient client.WithWatch, sharedCache ctrlcache.Cache, bufferSize int, numWorkers int, runClient workflowconnect.InternalRunServiceClient, recordFilterSize int, scope promutils.Scope) *ActionsClient {
	if numWorkers <= 0 {
		numWorkers = 1
	}
	c := &ActionsClient{
		k8sClient:   k8sClient,
		sharedCache: sharedCache,
		bufferSize:  bufferSize,
		numWorkers:  numWorkers,
		runClient:   runClient,
		subscribers: make(map[string]map[chan *ActionUpdate]struct{}),
	}

	if recordFilterSize > 0 {
		filter, err := fastcheck.NewOppoBloomFilter(recordFilterSize, scope.NewSubScope("actions_filter"))
		if err != nil {
			logger.Warnf(context.Background(), "Failed to create record filter (size=%d): %v; proceeding without dedup", recordFilterSize, err)
		} else {
			c.recordedFilter = filter
		}
	}

	return c
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
		taskActionName := buildTaskActionName(actionID)
		if err := k8sutil.EnsureNamespaceExists(ctx, c.k8sClient, flyteNamespace); err != nil {
			return fmt.Errorf("failed to ensure namespace %s: %w", flyteNamespace, err)
		}
		taskAction := &executorv1.TaskAction{
			ObjectMeta: metav1.ObjectMeta{
				Name:      taskActionName,
				Namespace: flyteNamespace,
				Labels: map[string]string{
					"flyte.org/project":     actionID.Run.Project,
					"flyte.org/domain":      actionID.Run.Domain,
					"flyte.org/run":         actionID.Run.Name,
					"flyte.org/action":      actionID.Name,
					"flyte.org/action-type": "task", // TODO: derive from action.Spec
					"flyte.org/is-root":     fmt.Sprintf("%t", isRoot),
				},
			},
			Spec: executorv1.TaskActionSpec{},
		}
		var parentTaskAction *executorv1.TaskAction
		// Set OwnerReference to parent so K8s cascades deletion to children.
		if !isRoot {
			parentID := &common.ActionIdentifier{
				Run:  actionID.Run,
				Name: *action.ParentActionName,
			}
			parentName := buildTaskActionName(parentID)

			parent := &executorv1.TaskAction{}
			if err := c.k8sClient.Get(ctx, client.ObjectKey{Name: parentName, Namespace: flyteNamespace}, parent); err != nil {
				return fmt.Errorf("failed to get parent TaskAction %s: %w", parentName, err)
			}
			parentTaskAction = parent

			blockOwnerDeletion := true
			taskAction.OwnerReferences = []metav1.OwnerReference{
				{
					APIVersion:         "flyte.org/v1",
					Kind:               "TaskAction",
					Name:               parent.Name,
					UID:                parent.UID,
					BlockOwnerDeletion: &blockOwnerDeletion,
				},
			}
			// For child actions, inherit parent's run context
			inheritRunContextFromParentTaskAction(taskAction, parentTaskAction)
		} else {
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
		if err := embedTaskTemplate(action, taskAction); err != nil {
			return fmt.Errorf("failed to embed task template: %w", err)
		}

		if err := c.k8sClient.Create(ctx, taskAction); err != nil {
			return fmt.Errorf("failed to create TaskAction CR %s: %w", taskActionName, err)
		}

		logger.Infof(ctx, "Created TaskAction CR: %s", taskActionName)
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
	if err := c.k8sClient.Get(ctx, client.ObjectKey{Name: taskActionName, Namespace: flyteNamespace}, taskAction); err != nil {
		return fmt.Errorf("failed to get TaskAction %s: %w", taskActionName, err)
	}

	if err := c.k8sClient.Delete(ctx, taskAction); err != nil {
		return fmt.Errorf("failed to delete TaskAction %s: %w", taskActionName, err)
	}

	logger.Infof(ctx, "Deleted TaskAction %s (descendants will be cascade deleted by K8s)", taskActionName)
	return nil
}

// GetState retrieves the state JSON for a TaskAction
func (c *ActionsClient) GetState(ctx context.Context, actionID *common.ActionIdentifier) (string, error) {
	taskActionName := buildTaskActionName(actionID)

	taskAction := &executorv1.TaskAction{}
	if err := c.k8sClient.Get(ctx, client.ObjectKey{
		Name:      taskActionName,
		Namespace: flyteNamespace,
	}, taskAction); err != nil {
		return "", fmt.Errorf("failed to get TaskAction %s: %w", taskActionName, err)
	}

	return taskAction.Status.StateJSON, nil
}

// PutState updates the state JSON and latest attempt metadata for a TaskAction.
func (c *ActionsClient) PutState(ctx context.Context, actionID *common.ActionIdentifier, attempt uint32, status *workflow.ActionStatus, stateJSON string) error {
	taskActionName := buildTaskActionName(actionID)

	// Get current TaskAction
	taskAction := &executorv1.TaskAction{}
	if err := c.k8sClient.Get(ctx, client.ObjectKey{
		Name:      taskActionName,
		Namespace: flyteNamespace,
	}, taskAction); err != nil {
		return fmt.Errorf("failed to get TaskAction %s: %w", taskActionName, err)
	}

	// Skip update if the stateJSON does not change
	if taskAction.Status.StateJSON == stateJSON {
		return nil
	}

	// Update state JSON
	taskAction.Status.StateJSON = stateJSON
	if status != nil {
		taskAction.Status.Attempts = status.GetAttempts()
		taskAction.Status.CacheStatus = status.GetCacheStatus()
	}

	// Update status subresource
	if err := c.k8sClient.Status().Update(ctx, taskAction); err != nil {
		return fmt.Errorf("failed to update TaskAction status %s: %w", taskActionName, err)
	}

	logger.Infof(ctx, "Updated state for TaskAction: %s", taskActionName)
	return nil
}

// ListRunActions lists all TaskActions belonging to a run.
func (c *ActionsClient) ListRunActions(ctx context.Context, runID *common.RunIdentifier) ([]*executorv1.TaskAction, error) {
	taskActionList := &executorv1.TaskActionList{}
	listOpts := []client.ListOption{
		client.InNamespace(flyteNamespace),
		client.MatchingLabels{
			"flyte.org/project": runID.Project,
			"flyte.org/domain":  runID.Domain,
			"flyte.org/run":     runID.Name,
		},
	}

	if err := c.k8sClient.List(ctx, taskActionList, listOpts...); err != nil {
		return nil, fmt.Errorf("failed to list TaskActions for run: %w", err)
	}

	result := make([]*executorv1.TaskAction, len(taskActionList.Items))
	for i := range taskActionList.Items {
		result[i] = &taskActionList.Items[i]
	}
	return result, nil
}

// ListChildActions lists all TaskActions that are children of the given parent action
func (c *ActionsClient) ListChildActions(ctx context.Context, parentActionID *common.ActionIdentifier) ([]*executorv1.TaskAction, error) {
	// List all TaskActions in the same run
	taskActionList := &executorv1.TaskActionList{}
	listOpts := []client.ListOption{
		client.InNamespace(flyteNamespace),
		client.MatchingLabels{
			"flyte.org/project": parentActionID.Run.Project,
			"flyte.org/domain":  parentActionID.Run.Domain,
			"flyte.org/run":     parentActionID.Run.Name,
		},
	}

	if err := c.k8sClient.List(ctx, taskActionList, listOpts...); err != nil {
		return nil, fmt.Errorf("failed to list TaskActions: %w", err)
	}

	// Filter for the parent and its children
	var result []*executorv1.TaskAction
	for i := range taskActionList.Items {
		action := &taskActionList.Items[i]
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
	if err := c.k8sClient.Get(ctx, client.ObjectKey{
		Name:      taskActionName,
		Namespace: flyteNamespace,
	}, taskAction); err != nil {
		return nil, fmt.Errorf("failed to get TaskAction %s: %w", taskActionName, err)
	}

	return taskAction, nil
}

// Subscribe creates a new subscription channel for action updates for specified parent action name
func (c *ActionsClient) Subscribe(parentActionName string) chan *ActionUpdate {
	c.mu.Lock()
	defer c.mu.Unlock()

	ch := make(chan *ActionUpdate, c.bufferSize)
	if c.subscribers[parentActionName] == nil {
		c.subscribers[parentActionName] = make(map[chan *ActionUpdate]struct{})
	}
	c.subscribers[parentActionName][ch] = struct{}{}
	return ch
}

// Unsubscribe removes the given channel from the subscription list for the parent action name
func (c *ActionsClient) Unsubscribe(parentActionName string, ch chan *ActionUpdate) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if channels, ok := c.subscribers[parentActionName]; ok {
		delete(channels, ch)
		close(ch)
		if len(channels) == 0 {
			delete(c.subscribers, parentActionName)
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
	for i := range c.workerChs {
		c.workerChs[i] = make(chan watch.Event, c.bufferSize)
		go c.worker(ctx, c.workerChs[i], stopCh)
	}
	c.mu.Unlock()

	logger.Infof(ctx, "Starting TaskAction watcher for namespace: %s (workers: %d)", flyteNamespace, c.numWorkers)

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
// TaskAction, using FNV-32a hashing for consistent sharding.
func (c *ActionsClient) dispatchEvent(taskAction *executorv1.TaskAction, eventType watch.EventType) {
	if taskAction == nil {
		return
	}

	h := fnv.New32a()
	_, _ = h.Write([]byte(taskAction.Name)) // FNV Write never returns an error
	shard := h.Sum32() % uint32(c.numWorkers)
	c.workerChs[shard] <- watch.Event{Type: eventType, Object: taskAction.DeepCopy()}
}

func (c *ActionsClient) handleWatchEvent(ctx context.Context, event watch.Event) {
	taskAction, ok := event.Object.(*executorv1.TaskAction)
	if !ok {
		logger.Warnf(ctx, "received non-TaskAction object in watch event: %T", event.Object)
		return
	}

	c.handleTaskActionEvent(ctx, taskAction, event.Type)
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

	// Determine short name: use spec.ShortName if set, otherwise extract from template ID
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
		OutputUri:        buildOutputUri(ctx, taskAction),
		IsDeleted:        eventType == watch.Deleted,
		TaskType:         taskAction.Spec.TaskType,
		ShortName:        shortName,
		ErrorState:       taskAction.Status.ErrorState,
	}
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

	for ch := range c.subscribers[update.ParentActionName] {
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

	// On ADDED: create the action record in the DB (deduplicated via bloom filter).
	if eventType == watch.Added {
		actionKey := []byte(buildTaskActionName(update.ActionID))
		isDuplicate := c.recordedFilter != nil && c.recordedFilter.Contains(ctx, actionKey)
		if isDuplicate {
			logger.Debugf(ctx, "Skipping duplicate RecordAction for %s", update.ActionID.Name)
		} else {
			recordReq := &workflow.RecordActionRequest{
				ActionId: update.ActionID,
				Parent:   update.ParentActionName,
				InputUri: taskAction.Spec.InputURI,
				Group:    taskAction.Spec.Group,
			}
			if taskAction.Spec.TaskType != "" {
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
			} else if c.recordedFilter != nil {
				c.recordedFilter.Add(ctx, actionKey)
			}
		}

		// When a child action appears, the parent must already be running (it
		// created the child). Promote the parent to RUNNING so the UI doesn't
		// stay stuck on INITIALIZING while children are executing.
		if !isDuplicate && update.ParentActionName != "" {
			parentID := &common.ActionIdentifier{
				Run:  update.ActionID.Run,
				Name: update.ParentActionName,
			}
			parentStatusReq := &workflow.UpdateActionStatusRequest{
				ActionId: parentID,
				Status: &workflow.ActionStatus{
					Phase: common.ActionPhase_ACTION_PHASE_RUNNING,
				},
			}
			if _, err := c.runClient.UpdateActionStatus(ctx, connect.NewRequest(parentStatusReq)); err != nil {
				logger.Warnf(ctx, "Failed to promote parent action %s to RUNNING: %v", update.ParentActionName, err)
			}
		}
	}

	if update.Phase != common.ActionPhase_ACTION_PHASE_UNSPECIFIED {
		statusReq := &workflow.UpdateActionStatusRequest{
			ActionId: update.ActionID,
			Status: &workflow.ActionStatus{
				Phase:       update.Phase,
				Attempts:    taskAction.Status.Attempts,
				CacheStatus: taskAction.Status.CacheStatus,
			},
		}
		if _, err := c.runClient.UpdateActionStatus(ctx, connect.NewRequest(statusReq)); err != nil {
			logger.Warnf(ctx, "Failed to update action status in run service for %s: %v", update.ActionID.Name, err)
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
		phase == common.ActionPhase_ACTION_PHASE_TIMED_OUT
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

// buildOutputUri computes the action-specific output URI from the TaskAction spec.
// It uses the same path structure as the executor's ComputeActionOutputPath so that
// the SDK can find outputs written by the executor.
func buildOutputUri(ctx context.Context, ta *executorv1.TaskAction) string {
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
		return
	}

	taskAction.Spec.EnvVars = keyValuePairsToMap(runSpec.GetEnvs().GetValues())
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

// embedTaskTemplate serializes the inline TaskTemplate from the Action into the CR spec.
func embedTaskTemplate(action *actions.Action, taskAction *executorv1.TaskAction) error {
	taskSpec, ok := action.Spec.(*actions.Action_Task)
	if !ok || taskSpec.Task == nil || taskSpec.Task.Spec == nil || taskSpec.Task.Spec.TaskTemplate == nil {
		// Non-task actions do not carry an inline template.
		return nil
	}

	tmpl := taskSpec.Task.Spec.TaskTemplate
	taskAction.Spec.TaskType = tmpl.Type
	taskAction.Spec.ShortName = taskSpec.Task.Spec.ShortName

	data, err := proto.Marshal(tmpl)
	if err != nil {
		return fmt.Errorf("failed to marshal task template: %w", err)
	}
	taskAction.Spec.TaskTemplate = data
	return nil
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
