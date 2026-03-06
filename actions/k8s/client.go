package k8s

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	executorv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
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
}

// ActionsClient handles all etcd/K8s TaskAction CR operations for the Actions service.
type ActionsClient struct {
	k8sClient  client.WithWatch
	namespace  string
	bufferSize int
	runClient  workflowconnect.InternalRunServiceClient

	// Watch management
	mu sync.RWMutex
	// Map parent action name to subscriber channels.
	// Multiple callers may watch the same parent action concurrently.
	// TODO: add a prometheus counter for dropped updates when metrics are wired up
	subscribers map[string]map[chan *ActionUpdate]struct{}
	stopCh      chan struct{}
	watching    bool
}

// NewActionsClient creates a new Kubernetes-based actions client.
func NewActionsClient(k8sClient client.WithWatch, namespace string, bufferSize int, runClient workflowconnect.InternalRunServiceClient) *ActionsClient {
	return &ActionsClient{
		k8sClient:   k8sClient,
		namespace:   namespace,
		bufferSize:  bufferSize,
		runClient:   runClient,
		subscribers: make(map[string]map[chan *ActionUpdate]struct{}),
	}
}

// Enqueue creates a TaskAction CR in etcd (via the K8s API).
func (c *ActionsClient) Enqueue(ctx context.Context, action *actions.Action, runSpec *task.RunSpec) error {
	actionID := action.ActionId
	logger.Infof(ctx, "Enqueuing action: %s/%s/%s/%s/%s",
		actionID.Run.Org, actionID.Run.Project, actionID.Run.Domain,
		actionID.Run.Name, actionID.Name)

	isRoot := action.ParentActionName == nil || *action.ParentActionName == ""

	taskActionName := buildTaskActionName(actionID)
	taskAction := &executorv1.TaskAction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      taskActionName,
			Namespace: c.namespace,
			Labels: map[string]string{
				"flyte.org/org":         actionID.Run.Org,
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

	// Set OwnerReference to parent so K8s cascades deletion to children.
	if !isRoot {
		parentID := &common.ActionIdentifier{
			Run:  actionID.Run,
			Name: *action.ParentActionName,
		}
		parentName := buildTaskActionName(parentID)

		parent := &executorv1.TaskAction{}
		if err := c.k8sClient.Get(ctx, client.ObjectKey{Name: parentName, Namespace: c.namespace}, parent); err != nil {
			return fmt.Errorf("failed to get parent TaskAction %s: %w", parentName, err)
		}

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
	}

	// Build and set the ActionSpec for the executor.
	actionSpec := buildActionSpec(action, runSpec)
	if err := taskAction.Spec.SetActionSpec(actionSpec); err != nil {
		return fmt.Errorf("failed to set action spec: %w", err)
	}

	// Embed the inline TaskTemplate if present.
	if err := embedTaskTemplate(action, taskAction); err != nil {
		return fmt.Errorf("failed to embed task template: %w", err)
	}

	if err := c.k8sClient.Create(ctx, taskAction); err != nil {
		return fmt.Errorf("failed to create TaskAction CR %s: %w", taskActionName, err)
	}

	logger.Infof(ctx, "Created TaskAction CR: %s", taskActionName)
	return nil
}

// AbortAction deletes a TaskAction CR from etcd.
// K8s cascades the deletion to all descendants via OwnerReferences.
func (c *ActionsClient) AbortAction(ctx context.Context, actionID *common.ActionIdentifier, reason *string) error {
	taskActionName := buildTaskActionName(actionID)
	logger.Infof(ctx, "Aborting action %s (reason: %v)", taskActionName, reason)

	taskAction := &executorv1.TaskAction{}
	if err := c.k8sClient.Get(ctx, client.ObjectKey{Name: taskActionName, Namespace: c.namespace}, taskAction); err != nil {
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
		Namespace: c.namespace,
	}, taskAction); err != nil {
		return "", fmt.Errorf("failed to get TaskAction %s: %w", taskActionName, err)
	}

	return taskAction.Status.StateJSON, nil
}

// PutState updates the state JSON for a TaskAction.
// attempt and status are accepted for future use (e.g. recording to RunService).
func (c *ActionsClient) PutState(ctx context.Context, actionID *common.ActionIdentifier, attempt uint32, status *workflow.ActionStatus, stateJSON string) error {
	taskActionName := buildTaskActionName(actionID)

	// Get current TaskAction
	taskAction := &executorv1.TaskAction{}
	if err := c.k8sClient.Get(ctx, client.ObjectKey{
		Name:      taskActionName,
		Namespace: c.namespace,
	}, taskAction); err != nil {
		return fmt.Errorf("failed to get TaskAction %s: %w", taskActionName, err)
	}

	// Skip update if the stateJSON does not change
	if taskAction.Status.StateJSON == stateJSON {
		return nil
	}

	// Update state JSON
	taskAction.Status.StateJSON = stateJSON

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
		client.InNamespace(c.namespace),
		client.MatchingLabels{
			"flyte.org/org":     runID.Org,
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
		client.InNamespace(c.namespace),
		client.MatchingLabels{
			"flyte.org/org":     parentActionID.Run.Org,
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
		Namespace: c.namespace,
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

// StartWatching starts watching TaskAction resources and notifies all subscribers
func (c *ActionsClient) StartWatching(ctx context.Context) error {
	c.mu.Lock()
	if c.watching {
		c.mu.Unlock()
		return nil
	}
	c.watching = true
	c.stopCh = make(chan struct{})
	c.mu.Unlock()

	logger.Infof(ctx, "Starting TaskAction watcher for namespace: %s", c.namespace)

	go c.watchLoop(ctx)

	return nil
}

// watchLoop continuously watches TaskAction resources
func (c *ActionsClient) watchLoop(ctx context.Context) {
	for {
		select {
		case <-c.stopCh:
			logger.Infof(ctx, "TaskAction watcher stopped")
			return
		case <-ctx.Done():
			logger.Infof(ctx, "TaskAction watcher context cancelled")
			return
		default:
			if err := c.doWatch(ctx); err != nil {
				logger.Warnf(ctx, "Watch error, will retry: %v", err)
				time.Sleep(5 * time.Second)
			}
		}
	}
}

// doWatch performs a single watch operation
func (c *ActionsClient) doWatch(ctx context.Context) error {
	taskActionList := &executorv1.TaskActionList{}

	watcher, err := c.k8sClient.Watch(ctx, taskActionList, client.InNamespace(c.namespace))
	if err != nil {
		return fmt.Errorf("failed to start watch: %w", err)
	}
	defer watcher.Stop()

	for {
		select {
		case <-c.stopCh:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return fmt.Errorf("watch channel closed")
			}
			c.handleWatchEvent(ctx, event)
		}
	}
}

// handleWatchEvent processes a watch event
func (c *ActionsClient) handleWatchEvent(ctx context.Context, event watch.Event) {
	taskAction, ok := event.Object.(*executorv1.TaskAction)
	if !ok {
		logger.Warnf(ctx, "Received non-TaskAction object in watch event")
		return
	}

	var parentName string
	if taskAction.Spec.ParentActionName != nil {
		parentName = *taskAction.Spec.ParentActionName
	}

	// Determine short name: use spec.ShortName if set, otherwise extract from template ID
	shortName := taskAction.Spec.ShortName
	if shortName == "" && len(taskAction.Spec.TaskTemplate) > 0 {
		shortName = extractShortNameFromTemplate(taskAction.Spec.TaskTemplate)
	}

	update := &ActionUpdate{
		ActionID: &common.ActionIdentifier{
			Run: &common.RunIdentifier{
				Org:     taskAction.Spec.Org,
				Project: taskAction.Spec.Project,
				Domain:  taskAction.Spec.Domain,
				Name:    taskAction.Spec.RunName,
			},
			Name: taskAction.Spec.ActionName,
		},
		ParentActionName: parentName,
		StateJSON:        taskAction.Status.StateJSON,
		Phase:            GetPhaseFromConditions(taskAction),
		OutputUri:        buildOutputUri(taskAction),
		IsDeleted:        event.Type == watch.Deleted,
		TaskType:         taskAction.Spec.TaskType,
		ShortName:        shortName,
	}

	c.notifySubscribers(ctx, update)
	go c.notifyRunService(ctx, taskAction, update, event.Type)
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
// On all events it calls UpdateActionStatus (when phase is meaningful) and RecordActionEvents.
func (c *ActionsClient) notifyRunService(ctx context.Context, taskAction *executorv1.TaskAction, update *ActionUpdate, eventType watch.EventType) {
	if c.runClient == nil {
		return
	}

	// On ADDED: create the action record in the DB.
	if eventType == watch.Added {
		recordReq := &workflow.RecordActionRequest{
			ActionId: update.ActionID,
			Parent:   update.ParentActionName,
			InputUri: taskAction.Spec.InputURI,
		}
		if taskAction.Spec.TaskType != "" {
			recordReq.Spec = &workflow.RecordActionRequest_Task{
				Task: &workflow.TaskAction{},
			}
		}
		if _, err := c.runClient.RecordAction(ctx, connect.NewRequest(recordReq)); err != nil {
			logger.Warnf(ctx, "Failed to record action in run service for %s: %v", update.ActionID.Name, err)
		}
	}

	if update.Phase != common.ActionPhase_ACTION_PHASE_UNSPECIFIED {
		statusReq := &workflow.UpdateActionStatusRequest{
			ActionId: update.ActionID,
			Status: &workflow.ActionStatus{
				Phase: update.Phase,
			},
		}
		if _, err := c.runClient.UpdateActionStatus(ctx, connect.NewRequest(statusReq)); err != nil {
			logger.Warnf(ctx, "Failed to update action status in run service for %s: %v", update.ActionID.Name, err)
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

// buildTaskActionName generates a Kubernetes-compliant name for the TaskAction
func buildTaskActionName(actionID *common.ActionIdentifier) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		actionID.Run.Org,
		actionID.Run.Project,
		actionID.Run.Domain,
		actionID.Run.Name,
		actionID.Name,
	)
}

// buildOutputUri computes the action-specific output URI from the TaskAction spec.
func buildOutputUri(ta *executorv1.TaskAction) string {
	if ta.Spec.RunOutputBase == "" {
		return ""
	}
	return strings.TrimRight(ta.Spec.RunOutputBase, "/") + "/" + ta.Spec.ActionName
}

// InitScheme adds the executor API types to the scheme
func InitScheme() error {
	return executorv1.AddToScheme(scheme.Scheme)
}

// buildActionSpec converts an actions.Action into the workflow.ActionSpec expected by the executor.
func buildActionSpec(action *actions.Action, runSpec *task.RunSpec) *workflow.ActionSpec {
	actionSpec := &workflow.ActionSpec{
		ActionId:      action.ActionId,
		ParentActionName: action.ParentActionName,
		RunSpec:       runSpec,
		InputUri:      action.InputUri,
		RunOutputBase: action.RunOutputBase,
		Group:         action.Group,
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
